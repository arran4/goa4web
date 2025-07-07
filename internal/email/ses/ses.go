package ses

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider wraps the AWS SES client.
type Provider struct {
	Client sesiface.SESAPI
	From   string
}

func (s Provider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	dest := &ses.Destination{ToAddresses: []*string{aws.String(to)}}
	msg := &ses.Message{
		Subject: &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(subject)},
		Body:    &ses.Body{},
	}
	msg.Body.Text = &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(textBody)}
	if htmlBody != "" {
		msg.Body.Html = &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(htmlBody)}
	}
	input := &ses.SendEmailInput{Destination: dest, Message: msg, Source: aws.String(s.From)}
	_, err := s.Client.SendEmailWithContext(ctx, input)
	return err
}

func providerFromConfig(cfg runtimeconfig.RuntimeConfig) email.Provider {
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
	return Provider{Client: ses.New(sess), From: cfg.EmailFrom}
}

// Register registers the SES provider factory.
func Register() { email.RegisterProvider("ses", providerFromConfig) }
