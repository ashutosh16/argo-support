# ibp-genai-service

[![Build Status](https://build.intuit.com/devx-shared//buildStatus/buildIcon?job=dev-build/ibp-genai-service/ibp-genai-service/master)](https://build.intuit.com/devx-shared//job/dev-build/job/ibp-genai-service/job/ibp-genai-service/job/master/)
[![Code Coverage](https://codecov.tools.a.intuit.com/ghe/dev-build/ibp-genai-service/branch/master/graph/badge.svg)](https://codecov.tools.a.intuit.com/ghe/dev-build/ibp-genai-service)
[![SSDLC Badge][ssdlc-image]][ssdlc-url]

[ssdlc-image]: https://badge.ssdlc.a.intuit.com/ssdlc/coverage/1732777549890492173
[ssdlc-url]: https://devportal.intuit.com/app/dp/resource/1732777549890492173/security/security

DevPortal Asset
------
https://devportal.intuit.com/app/dp/resource/1732777549890492173/configuration/environment

# Go && `github.intuit.com`

To use Go with `github.intuit.com`, you need to follow these steps to configure a `github.intuit.com` token.

- Generate a new github token here: https://github.intuit.com/settings/tokens/new?scopes=repo
- Assign your token as an environment variable: `export GITHUB_INTUIT_TOKEN=INSERT_TOKEN_HERE`
- Use this command to update your `~/.gitconfig` file to use the token to authenticate to github.intuit.com
    - `git config --global --add url."https://${GITHUB_INTUIT_TOKEN}@github.intuit.com".insteadOf "https://github.intuit.com"`
- Add this to your `~/.bash_profile` file
    - `export GOPRIVATE=github.intuit.com`

# Local Development

## Prerequisites

- Install Go: `brew install go`
- Install golangci-lint: `brew install golangci-lint`

## Configuring your environment

### Run the service against PreProd (E2E)

The services requires the following environment variables to be set in order to run against PreProd (E2E):
* It will use the PreProd (E2E) IDPS endpoint to retrieve the [Identity App Secret](https://devportal.intuit.com/app/dp/resource/1732777549890492173/credentials/secrets)
* It will use the PreProd (E2E) Identity endpoint to retrieve the Authentication Header required to consume the 
PreProd (E2E) Express endpoint using the PreProd (E2E) [Offline JobID](https://devportal.intuit.com/app/dp/resource/1732777549890492173/credentials/offlineJob)
* It will submit GenAI request to the PreProd (E2E) Express endpoint

```
### PreProd (QAL and E2E)
# Express Endpoint
export APP_EXPRESS_ENDPOINT="https://genpluginregistry-e2e.api.intuit.com/v1/llmexpress"

# Identity configuration
export APP_IDENTITY_ENDPOINT="https://identityinternal-e2e.api.intuit.com/v1/graphql"
export APP_IDENTITY_JOB_ID="9341450931784620"

# IDPS configuration (to retrieve Identity App Secret)
export APP_IDPS_ENDPOINT="vkm-e2e.ps.idps.a.intuit.com"
export APP_IDPS_POLICY="p-iqv3508ask8u"
export APP_IDPS_FOLDER="ibpjenkins/e2e"

# Splunk configuration
export APP_SPLUNK_HOSTNAME="hec-us-west-2.e2e.cmn.cto.a.intuit.com"
export APP_SPLUNK_TOKEN=[see below]
```

The PreProd (E2E) `APP_IDPS_POLICY` above is a policy created by the IBP team for local development.
It is defined [here](https://devportal.intuit.com/app/dp/resource/8073845825132550131/addons/idps/manager).

> Note: Please reach out to the IBP team to get the APP_SPLUNK_TOKEN value.

### Run the service against Production

The services requires the following environment variables to be set in order to run against production:
* It will use the production IDPS endpoint to retrieve the [Identity App Secret](https://devportal.intuit.com/app/dp/resource/1732777549890492173/credentials/secrets)
* It will use the production Identity endpoint to retrieve the Authentication Header required to consume the
  production Express endpoint using the production [Offline JobID](https://devportal.intuit.com/app/dp/resource/1732777549890492173/credentials/offlineJob)
* It will submit GenAI request to the production Express endpoint

```
### Prod (STG and PROD)
# Express Endpoint
export APP_EXPRESS_ENDPOINT="https://genpluginregistry.api.intuit.com/v1/llmexpress"

# Identity configuration
export APP_IDENTITY_ENDPOINT="https://identityinternal.api.intuit.com/v1/graphql"
export APP_IDENTITY_JOB_ID="9341451835093753"

# IDPS configuration (to retrieve Identity App Secret)
export APP_IDPS_ENDPOINT="vkm.ps.idps.a.intuit.com"
export APP_IDPS_POLICY="p-rzc0yhkrbj3x"
export APP_IDPS_FOLDER="ibpjenkins/prod"

# Splunk configuration
export APP_SPLUNK_HOSTNAME="hec-us-west-2.e2e.cmn.cto.a.intuit.com"
export APP_SPLUNK_TOKEN=[see below]
```

The production `APP_IDPS_POLICY` above is a policy created by the IBP team for local development.
It is defined [here](https://devportal.intuit.com/app/dp/resource/8073845825132550131/addons/idps/manager).

> Note: Please reach out to the IBP team to get the APP_SPLUNK_TOKEN value.

## Enabling IDPS

### For PreProd (E2E)

IDPS relies on `eiamCli` to be installed and configured. In order for IDPS to work, you need to execute the following command:
```
eiamCli login
eiamCli aws_creds -a 6367-6488-2764 -r PowerUser -p tep-300-poweruser
export AWS_PROFILE=tep-300-poweruser
```

### For Production

IDPS relies on `eiamCli` to be installed and configured. In order for IDPS to work, you need to execute the following command:
```
eiamCli login
eiamCli aws_creds -a 7335-3620-4770 -r PowerUser -p ibp-100-poweruser
export AWS_PROFILE=ibp-100-poweruser
```

# Usage

## Running the service locally

- Run `make` or `make run` to run the service locally
- This will run the server in the development configuration on port `:8080`
- You can check the service is running by executing `curl localhost:8080/health/full`

You can test the log analyzer by running:
```
curl -v \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{
          "failures": [
            {
              "context": "docker push .\ninvalid reference format"
            }
          ]
        }' \
http://localhost:8080/v2/analyze
```

## Running the tests locally

- Run `make test` to run the tests locally

# Contributing

Please check the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information on how to contribute to this project.

# Support

Please reach out to the IBP team on the [#ibp-community](https://intuit.enterprise.slack.com/archives/C9YFBNJBV) Slack channel for any support.
