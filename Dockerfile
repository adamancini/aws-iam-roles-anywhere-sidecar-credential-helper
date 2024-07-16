ARG GO_VERSION=1.21.1
ARG ALPINE_VERSION=3.18
ARG APP_VERSION="v0.0.0-unknown"
FROM --platform=linux/amd64 golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build

COPY . /go/src/github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper
RUN cd /go/src/github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$APP_VERSION" -a -installsuffix cgo -o credential-helper .

ARG ALPINE_VERSION=3.20
ARG APP_VERSION="v0.0.0-unknown"
FROM --platform=linux/amd64 alpine:${ALPINE_VERSION} as release
ENV APP_VERSION=$APP_VERSION
ENV AWS_REGION=us-east-2
ENV AWS_DEFAULT_REGION=us-east-2
ENV AWS_REFRESH_INTERVAL=300
ENV AWS_CONTAINER_CREDENTIALS_FULL_URI=http://localhost:8080/creds
LABEL maintainer="ada mancini <ada@replicated.com>"
LABEL description="AWS IAM Anywhere credential writer"
LABEL repo="https://github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper"
LABEL author="ada mancini <ada@replicated.com>"
LABEL version=$APP_VERSION

COPY --from=build /go/src/github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper/credential-helper /usr/local/bin/credential-helper
RUN chmod +x /usr/local/bin/credential-helper

ENTRYPOINT [ "/usr/local/bin/credential-helper" ]

