//go:build ses
// +build ses

package ses

import (
	"context"
	"log"
	"net/mail"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Provider wraps the AWS SES client.
type Provider struct {
	Client sesiface.SESAPI
	From   string
}

// Built indicates whether the SES provider is compiled in.
const Built = true

func (s Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	input := &ses.SendRawEmailInput{
		Destinations: []*string{aws.String(to.Address)},
		Source:       aws.String(s.From),
		RawMessage:   &ses.RawMessage{Data: rawEmailMessage},
	}
	_, err := s.Client.SendRawEmailWithContext(ctx, input)
	return err
}

func providerFromConfig(cfg config.RuntimeConfig) email.Provider {
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
