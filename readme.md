# GOA4Web

This repository contains the source code for a collection of web services written in Go. The original project dates back to 2006.

## Email Provider Configuration

Email notifications can be sent via several backends. Set `EMAIL_PROVIDER` to select one of the following modes:

- `ses` (default): Amazon SES. Requires valid AWS credentials and `AWS_REGION`.
- `smtp`: Standard SMTP server using `SMTP_HOST`, optional `SMTP_PORT`, `SMTP_USER`, and `SMTP_PASS`.
- `local`: Uses the local `sendmail` binary.
- `jmap`: Sends mail using JMAP. Requires `JMAP_ENDPOINT`, `JMAP_USER`, `JMAP_PASS`,
  `JMAP_ACCOUNT`, and `JMAP_IDENTITY`.
- `log`: Writes emails to the application log.

If configuration or credentials are missing, email is disabled and a log message is printed.

