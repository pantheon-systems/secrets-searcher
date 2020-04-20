package structures

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "regexp"
)

type RegexpSet []*regexp.Regexp

func NewRegexpSet(values []*regexp.Regexp) (result RegexpSet) {
    result = RegexpSet{}
    for _, value := range values {
        result = append(result, value)
    }
    return
}

func NewRegexpSetFromStrings(values []string) (result RegexpSet, err error) {
    result = RegexpSet{}
    for _, value := range values {
        var re *regexp.Regexp
        re, err = regexp.Compile(value)
        if err != nil {
            err = errors.Wrapv(err, "unable to parse regex string", value)
            return
        }

        result = append(result, re)
    }

    return
}

func NewRegexpSetFromStringsMustCompile(values []string) (result RegexpSet) {
    var err error
    result, err = NewRegexpSetFromStrings(values)
    if err != nil {
        panic(err)
    }

    return
}

func (r RegexpSet) FindStringSubmatchAny(input string) (result []string) {
    for _, re := range r {
        matches := re.FindStringSubmatch(input)
        if matches == nil {
            continue
        }
        return matches
    }

    return
}

func (r RegexpSet) MatchAny(input string) (result bool) {
    for _, re := range r {
        if re.MatchString(input) {
            return true
        }
    }

    return false
}

// FIXME This doesn't belong here, it's too specific to this app
func (r RegexpSet) MatchAndTestSubmatchOrOverlap(input string, rangeToMatch *LineRange) (result bool) {
    for _, re := range r {
        matches := re.FindAllStringSubmatchIndex(input, -1)

        for _, match := range matches {
            matchLineRange := NewLineRange(match[0], match[1])

            // If there's a backreference, it's location should match the provided location
            if len(match) > 2 {
                backrefLineRange := NewLineRange(match[2], match[3])
                if backrefLineRange.Equals(rangeToMatch) {
                    return true
                }

                continue
            }

            // If no backreference but the match overlaps, return true
            if matchLineRange.Overlaps(rangeToMatch) {
                return true
            }
        }
    }

    return false
}
