FROM --platform=linux/amd64 golang:1.21.1-alpine3.18 as build
LABEL maintainer="ada mancini <ada@replicated.com>"
LABEL description="AWS IAM Anywhere credential writer"
LABEL author="ada mancini <ada@replicated.com>"

COPY . /go/src/github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper
RUN cd /go/src/github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o credential-helper .


FROM --platform=linux/amd64 alpine:3.18 as release
ENV AWS_REGION=us-east-2
ENV AWS_DEFAULT_REGION=us-east-2
ENV AWS_CONTAINER_CREDENTIALS_FULL_URI=http://localhost:8080/creds

COPY --from=build /go/src/github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper/credential-helper /usr/local/bin/credential-helper
RUN chmod +x /usr/local/bin/credential-helper

ENTRYPOINT [ "/usr/local/bin/credential-helper" ]

