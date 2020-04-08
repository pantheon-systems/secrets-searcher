# search-secrets

Search for sensitive information stored in one or more git repositories.

## Install

To build, run:

```shell script
make
```

## Pantheon usage

Obtain a GitHub token that has `repo` access to the repositories you want to scan. To generate a new token, log into
GitHub, then go to "Settings > Developer settings > Personal access tokens > Generate new token".

Run this way:

```shell script
cd pantheon
../search-secrets --config="config.yaml" --log-level="debug" --source.api-token="$GITHUB_TOKEN"
```

The tool will create an `output/report` directory that includes an HTML report, and a collection of YAML files, one for
each secret that was found.

### Whitelisting

Whitelisting secrets can be done by matching the file path or the line of code using regular expressions, or the secret
can be whitelisted directly.

We should only whitelist secrets directly if it is an actual secret, valid from a business standpoint, that has been
evaluated and deemed safe.

It should not be used for false positives, like an entropy rule matching a public key string. In that case, it would
be more appropriate to whitelist the secret by matching the line of code it was found in.

Or if an entropy rule is reporting individual lines from PEM files, but the "pem" rules are already reporting those PEM
files. That could be considered a bug, and it would be more appropriate to fix tool itself.

If we follow this guideline, the `./pantheon/whitelist` directory will remain a useful, self-documenting log of committed
secrets that were remediated, and our rule configuration will be improved over time.

#### Whitelisting secrets by matching the path of the file they are found in

In `./pantheon/config.yaml`, add an entry to the `whitelist-path-match` list.

#### Whitelisting secrets by matching the line of code they are found in

In `./pantheon/config.yaml`, add an entry to the `rules.[n].whitelist-code-match` list on the appropriate rule.

#### Whitelisting secrets directly

Copy the secret's JSON file from `./report/secrets` to `./pantheon/whitelist` and commit your changes. Add your comments to
the file.

Note: This type of whitelisting should only be used for actual secrets, valid from a business standpoint, that have been
evaluated and deemed safe. See "Whitelisting secrets" section above.
