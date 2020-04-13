package processor

import (
    "github.com/pantheon-systems/search-secrets/pkg/entropy"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    urlpkg "net/url"
    "regexp"
    "strings"
)

const (
    // FIXME These belong in config
    secretInPathLengthThreshold  = 20
    secretInPathEntropyThreshold = 4.5
)

var (
    // This finds possible URLs in the line
    findPossibleURLsInStringRe = regexp.MustCompile(`(?i)(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s\x60!()\[\]{};:'".,<>?«»“”‘’]))`)

    skipURLStringRes = structures.NewRegexpSetFromStringsMustCompile([]string{
        `^-`,
    })

    // TODO: url.Parse chokes on "http://username:{PASS}@..." template variables so maybe use this somehow
    // and keep checking the URL. Right now, the whole URL is skipped with a warning.
    templatePasswordInURLRe     = regexp.MustCompile(`[a-zA-Z]{3,10}://[^:]+:([^@]+)@`)
    templateVariablePasswordRes = structures.NewRegexpSetFromStringsMustCompile([]string{
        // Matches "$DBPASS", "$password", etc
        `(?i)^\$[a-z_]*(?:token|pass(?:word))?$`,
        // Matches "{DBPASS}", "{password}", etc
        `(?i)^{[a-z_]*(?:token|pass(?:word))?}$`,
    })
)

type URLProcessor struct {
    whitelistCodeRes *structures.RegexpSet
}

func NewURLProcessor(whitelistCodeRes *structures.RegexpSet) (result *URLProcessor) {
    result = &URLProcessor{
        whitelistCodeRes: whitelistCodeRes,
    }

    return
}

func (p *URLProcessor) FindInFileChange(*rule.FileChangeContext, *logrus.Entry) (result []*rule.FileChangeFinding, ignore []*structures.FileRange, err error) {
    return
}

func (p *URLProcessor) FindInLine(line string, log *logrus.Entry) (result []*rule.LineFinding, ignore []*structures.LineRange, err error) {
    indexPairs := findPossibleURLsInStringRe.FindAllStringIndex(line, -1)

    for _, pair := range indexPairs {
        lineRange := structures.NewLineRange(pair[0], pair[1])
        urlString := lineRange.ExtractValue(line).Value

        if skipURLStringRes.MatchAny(urlString) {
            continue
        }

        // Parse URL
        url, parseErr := urlpkg.Parse(urlString)
        if parseErr != nil {
            errors.ErrorLogForEntry(log, parseErr).Error("regex matched URL but url.Parse() could not, dropping match")
            continue
        }

        // Send ignore
        ignore = append(ignore, lineRange)

        secrets, findErr := p.findSecretsInURL(url, urlString, lineRange.StartIndex, log)
        if findErr != nil {
            errors.ErrorLogForEntry(log, findErr).Error("unable to find secrets in URL, dropping match")
            continue
        }
        if secrets == nil {
            continue
        }

        // Filter out whitelisted secrets
        secrets = p.filterWhitelistedSecrets(line, secrets)
        if secrets == nil {
            continue
        }

        for _, secret := range secrets {
            result = append(result, &rule.LineFinding{
                LineRange: secret.LineRange,
                Secrets:   []*rule.Secret{{Value: secret.Value}},
            })
        }
    }

    return
}

func (p *URLProcessor) findSecretsInURL(url *urlpkg.URL, urlString string, urlStartIndex int, log *logrus.Entry) (result []*structures.LineRangeValue, err error) {
    // Check password in URL
    var password *structures.LineRangeValue
    password, err = p.findPasswordInURL(url, urlString, urlStartIndex)
    if err != nil {
        return
    }
    if password != nil {
        result = append(result, password)
    }

    // Find high entropy in path
    highEntropyWords := p.findHighEntropyWordsInURLPath(url.Path, urlString, urlStartIndex, log)
    if highEntropyWords != nil {
        result = append(result, highEntropyWords...)
    }

    return
}

func (p *URLProcessor) filterWhitelistedSecrets(line string, secrets []*structures.LineRangeValue) (result []*structures.LineRangeValue) {
    for _, secret := range secrets {
        if p.isSecretWhitelisted(line, secret) {
            continue
        }

        result = append(result, secret)
    }
    return
}

func (p *URLProcessor) findPasswordInURL(url *urlpkg.URL, urlString string, urlStartIndex int) (result *structures.LineRangeValue, err error) {
    if url.User == nil {
        return
    }

    password, passwordSet := url.User.Password()
    if !passwordSet || password == "" {
        return
    }
    if templateVariablePasswordRes.MatchAny(password) {
        return
    }

    // url.URL can't tell us where in the URL string the password is, so we need to figure that out
    passwordInURLRe := regexp.MustCompile(`[a-zA-Z]{3,10}://[^:]+:(` + regexp.QuoteMeta(password) + `)@`)
    matches := passwordInURLRe.FindStringSubmatchIndex(urlString)
    if matches == nil {
        err = errors.New("url.URL has a password but our regex can't find its location")
        return
    }
    startIndex := urlStartIndex + matches[2]
    endIndex := urlStartIndex + matches[3]

    result = structures.NewLineRange(startIndex, endIndex).NewValue(password)

    return
}

func (p *URLProcessor) findHighEntropyWordsInURLPath(urlPath string, urlString string, urlStartIndex int, log *logrus.Entry) (result []*structures.LineRangeValue) {
    if len(urlPath) < 5 {
        return
    }

    pathPieces := strings.Split(urlPath, "/")
    pathPiecesLen := len(pathPieces)

    pathStartIndex := strings.Index(urlString, urlPath)
    if pathStartIndex == -1 {
        log.WithField("urlString", urlString).WithField("urlPath", urlPath).
            Error("url.URL has a path but we can't find it in the original URL string")
        return
    }

    // Index of the URL in the line, plus the index of the path in the URL, plus 1 for the leading slash in the URL path
    startIndex := urlStartIndex + pathStartIndex
    for i, pathPiece := range pathPieces {
        pathPieceLen := len(pathPiece)
        hasEntropy := p.isURLPathPieceHighEntropy(pathPiece, log)

        if hasEntropy {
            endIndex := startIndex + pathPieceLen
            result = append(result, structures.NewLineRange(startIndex, endIndex).NewValue(pathPiece))
        }

        if i < pathPiecesLen-1 {
            // Add the length of the URL path piece, plus 1 for the new slash in the URL path
            startIndex += pathPieceLen + 1
        }
    }

    return
}

func (p *URLProcessor) isURLPathPieceHighEntropy(pathPiece string, log *logrus.Entry) (result bool) {
    pathPieceLen := len(pathPiece)

    if pathPieceLen == 0 || pathPieceLen < secretInPathLengthThreshold {
        return
    }

    hasEntropy, err := entropy.HasHighEntropy(pathPiece, entropy.Base64CharsetName, secretInPathEntropyThreshold)
    if err != nil {
        errors.ErrorLogForEntry(log, err).Error("unable to evaluate path piece for high entropy")
        return
    }

    result = hasEntropy

    return
}

func (p *URLProcessor) isSecretWhitelisted(line string, secret *structures.LineRangeValue) bool {
    return p.whitelistCodeRes != nil && p.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(line, secret.LineRange)
}