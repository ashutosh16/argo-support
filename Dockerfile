FROM docker.intuit.com/docker-rmt/golang:1.22.4-alpine3.19 as build

ARG GITHUB_INTUIT_TOKEN
ARG GIT_COMMIT
ARG VERSION

# The following ARG and 2 LABEL are used by Jenkinsfile command
# to identify this intermediate container, for extraction of
# code coverage and other reported values.
ARG build
LABEL build=${build}
LABEL image=build

ENV GOPRIVATE=github.intuit.com
ENV LAST_MILE_PATH=data-curation/go/v0.7.16/lastmile
ENV UTILS_PATH=data-curation/go/v0.7.16/intuit
ENV GOLINTER_VERSION=v1.58.1

RUN apk add --no-cache git curl make bash

WORKDIR /go/src/github.intuit.com/dev-build/ibp-genai-service

# allow go to pull depencenies from github.intuit.com trough the GITHUB_INTUIT_TOKEN enviroment variable
RUN git config --global --add url."https://${GITHUB_INTUIT_TOKEN}@github.intuit.com".insteadOf "https://github.intuit.com"
COPY go.mod go.mod
COPY go.sum go.sum

# by downloading dependencies before adding source code, the download can be cached for subsequent runs
RUN go mod download
RUN curl -sSL https://${GITHUB_INTUIT_TOKEN}@github.intuit.com/raw/${LAST_MILE_PATH}/ppd.pem -o ppd.pem
RUN curl -sSL https://${GITHUB_INTUIT_TOKEN}@github.intuit.com/raw/${LAST_MILE_PATH}/prd.pem -o prd.pem
RUN curl -sSL https://${GITHUB_INTUIT_TOKEN}@github.intuit.com/raw/${UTILS_PATH}/utils.sh -o utils.sh
RUN curl -sSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${GOLINTER_VERSION}
COPY ./ /go/src/github.intuit.com/dev-build/ibp-genai-service

# Lint
RUN golangci-lint run --timeout 10m

# Build
RUN CGO_ENABLED=0 go build -ldflags "-X main.Version=$VERSION -X main.SHA=$GIT_COMMIT" -o main cmd/ibp-genai-service/main.go

# this entry point is only used when running tests, see the Makefile for usage
RUN make test


# ---------------------------
FROM docker.intuit.com/docker-rmt/alpine:3.19.1
ARG GIT_COMMIT
ARG DOCKER_TAGS=latest
# ARG JIRA_PROJECT=https://jira.intuit.com/projects/<CHANGE_ME>

# Required
ARG DOCKER_IMAGE_NAME=docker.intuit.com/dev-build/ibp-genai-service/service/ibp-genai-service:${DOCKER_TAGS}
ARG SERVICE_LINK=https://devportal.intuit.com/app/dp/resource/1732777549890492173

# Required
LABEL maintainer=some_email@intuit.com \
      app=ibp-genai-service \
      app-scope=runtime

USER root

# Create the appuser and appuser group
RUN addgroup -S appuser && adduser -D -G appuser appuser

# Install jq and openssl
RUN apk add --no-cache jq openssl bash

COPY entrypoint.sh /home/appuser/entrypoint.sh
RUN chmod +x /home/appuser/entrypoint.sh

USER appuser
WORKDIR /home/appuser

COPY --from=build /go/src/github.intuit.com/dev-build/ibp-genai-service/filtered_coverage.out /coverage.out
COPY --from=build /go/src/github.intuit.com/dev-build/ibp-genai-service/main /home/appuser/main
COPY --from=build /go/src/github.intuit.com/dev-build/ibp-genai-service/*.pem /home/appuser/lastmile/
COPY --from=build /go/src/github.intuit.com/dev-build/ibp-genai-service/utils.sh /home/appuser/utils.sh

ENTRYPOINT ["/home/appuser/entrypoint.sh"]
