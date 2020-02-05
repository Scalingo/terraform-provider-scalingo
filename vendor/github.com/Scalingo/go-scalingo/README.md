[ ![Codeship Status for Scalingo/go-scalingo](https://app.codeship.com/projects/cf518dc0-0034-0136-d6b3-5a0245e77f67/status?branch=master)](https://app.codeship.com/projects/279805)

# Go client for Scalingo API v4.0.0

## Release a new version

Bump new version number in:

- `CHANGELOG.md`
- `README.md`
- `version.go`

Tag and release a new version on GitHub
[here](https://github.com/Scalingo/go-scalingo/releases/new).

## Mocks

Generate the mocks with:

```shell
for interface in $(grep --extended-regexp --no-message --no-filename "type (.*Service|API|TokenGenerator) interface" ./* | grep -v  mockgen | cut -d " " -f 2)
do
  echo "Generating mock for $interface"
  if [[ $interface != "SubresourceService" ]]; then
    mockgen -destination scalingomock/gomock_$(echo $interface | tr '[:upper:]' '[:lower:]').go -package scalingomock github.com/Scalingo/go-scalingo $interface
  else
    mockgen -destination gomock_$(echo $interface | tr '[:upper:]' '[:lower:]').go -package scalingo github.com/Scalingo/go-scalingo $interface
  fi
done
```
