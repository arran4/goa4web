//go:build ses
// +build ses

package email

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"

	"github.com/arran4/goa4web/runtimeconfig"
)

// SESBuilt indicates whether the SES provider is compiled in.
const SESBuilt = true

// SESProvider wraps the AWS SES client.
type SESProvider struct{ Client sesiface.SESAPI }

// Send delivers an email using AWS SES.
func (s SESProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	dest := &ses.Destination{ToAddresses: []*string{aws.String(to)}}
	msg := &ses.Message{
		Subject: &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(subject)},
		Body:    &ses.Body{},
	}
	msg.Body.Text = &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(textBody)}
	if htmlBody != "" {
		msg.Body.Html = &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(htmlBody)}
	}
	input := &ses.SendEmailInput{Destination: dest, Message: msg, Source: aws.String(SourceEmail)}
	_, err := s.Client.SendEmailWithContext(ctx, input)
	return err
}

// SESProviderFromConfig constructs an SESProvider using cfg.
func SESProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider {
	awsCfg := aws.NewConfig()
	if region := cfg.EmailAWSRegion; region != "" {
		awsCfg = awsCfg.WithRegion(region)
	}
	sess, err := session.NewSession(awsCfg)
	if err != nil {
		log.Printf("Email disabled: cannot initialise AWS session: %v", err)
		return nil
	}
	if _, err := sess.Config.Credentials.Get(); err != nil {
		log.Printf("Email disabled: no AWS credentials: %v", err)
		return nil
	}
	return SESProvider{Client: ses.New(sess)}
}
