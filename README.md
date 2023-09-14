# aws-iam-roles-anyhwere-sidecar-credential-helper

## Description

This is a sidecar container used to inject AWS IAM credentials/config files into any container that needs them. It is designed to be used with the [aws-iam-roles-anywhere-sidecar](https://github.com/josh23french/iam-roles-anywhere-sidecar) to support applications that expect filesystem credentials from `~/.aws/credentials` or `~/.aws/config` and are unable to obtain credentials from an HTTP request to `$AWS_CONTAINER_CREDENTIALS_FULL_URI`.

## Usage

`credential-helper` must be used in conjunction with `aws-iam-roles-anywhere-sidecar`.  It issues an HTTP GET request to the `$AWS_CONTAINER_CREDENTIALS_FULL_URI` advertised by the sidecar; by default this endpoint is at `localhost:8080/creds`.  The credentials endpoint returns a JSON response and credential-helper transforms the JSON into an .ini-formatted file and writes it atomically to `~/.aws/credentials` and `~/.aws/config`.  Today, the container runs as the `root` user.  The container will refresh the credentials files every `$AWS_REFRESH_INTERVAL` seconds by writing to a temp file and then renaming the temp file to the credentials file.  This means that the inode of the credentials file will change on every refresh, which may cause issues with applications that expect the inode to remain the same.

## Configuration

### Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| AWS_REGION | The AWS region to use for the credential helper | us-east-2 |
| AWS_DEFAULT_REGION | The AWS region to use for the credential helper | us-east-2 |
| AWS_REFRESH_INTERVAL | The interval in seconds to refresh the credentials | 900 |
| AWS_CONTAINER_CREDENTIALS_FULL_URI | The URI to use to obtain credentials | http://localhost:8080/creds |
| LISTEN_PORT | The port to listen on for HTTP requests. | 3000 |

### Kubernetes

```yaml
spec:
  template:
    metadata:
    spec:
      containers:
      # iam-roles-anywhere-sidecar
      - env:
        - name: PRIVATE_KEY_ID
          value: "/var/run/key.pem"
        - name: CERTIFICATE_ID
          value: "/var/run/certificate.pem"
        - name: AWS_REGION
          value: "us-east-2"
        - name: AWS_DEFAULT_REGION
          value: "us-east-2"
        - name: NO_VERIFY_SSL
          value: "false"
        - name: ROLE_ARN
          value: "arn:aws:iam::1234567890ab:role/development-role"
        - name: PROFILE_ARN
          value: "arn:aws:rolesanywhere:us-east-2:1234567890ab:profile/31123e3d-f033-49cd-b1b4-eecbbca4c123"
        - name: TRUST_ANCHOR_ID
          value: "arn:aws:rolesanywhere:us-east-2:1234567890ab:trust-anchor/eefg123d-62da-4297-b9cb-fefg12345679"
        - name: DEBUG
          value: "false"
        - name: CREDENTIAL_VERSION
          value: "1"
        - name: SESSION_DURATION
          value: "900"
        - name: LISTEN
          value: "0.0.0.0:8080"
        image: ghcr.io/josh23french/iam-roles-anywhere-sidecar:v0.0.1
        volumeMounts:
        - mountPath: /var/run/certificate.pem
          name: certificate-id
          readOnly: true
          subPath: certificate.pem
        - mountPath: /var/run/key.pem
          name: private-key-id
          readOnly: true
          subPath: key.pem
      # aws-consumer
      - env:
        image: aws-consumer:0.0.0
        volumeMounts:
        - mountPath: /root/.aws
          name: aws-credentials
          readOnly: true
      # aws-iam-roles-anywhere-sidecar-credential-helper
      - env:
        - name: AWS_DEFAULT_REGION
          value: us-east-2
        - name: AWS_CONTAINER_CREDENTIALS_FULL_URI
          value: http://localhost:8080/creds
        - name: AWS_REFRESH_INTERVAL
          value: "300"
        image: adamancini/aws-iam-roles-anywhere-sidecar-credential-helper:v1.0.1
        name: credential-helper
        volumeMounts:
        - mountPath: /root/.aws
          name: aws-credentials
      volumes:
      - name: aws-credentials
        emptyDir:
          sizeLimit: 10Mi
      - name: certificate-id
        secret:
          defaultMode: 420
          secretName: iam-anywhere-certificate
      - name: private-key-id
        secret:
          defaultMode: 420
          secretName: iam-anywhere-private-key
```
