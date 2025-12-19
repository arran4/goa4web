//go:build ses

package ses

import (
	"context"
	"fmt"
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

func (s Provider) TestConfig(ctx context.Context) error {
	_, err := s.Client.GetSendQuotaWithContext(ctx, &ses.GetSendQuotaInput{})
	if err != nil {
		return fmt.Errorf("failed to get send quota: %w", err)
	}
	fmt.Println("SES provider is configured correctly")
	return nil
}

func providerFromConfig(cfg *config.RuntimeConfig) (email.Provider, error) {
	awsCfg := aws.NewConfig()
	if region := cfg.EmailAWSRegion; region != "" {
		awsCfg = awsCfg.WithRegion(region)
	}
	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, fmt.Errorf("Email disabled: cannot initialise AWS session: %v", err)
	}
	if _, err := sess.Config.Credentials.Get(); err != nil {
		return nil, fmt.Errorf("Email disabled: no AWS credentials: %v", err)
	}
	return Provider{Client: ses.New(sess), From: cfg.EmailFrom}, nil
}

// Register registers the SES provider factory.
func Register(r *email.Registry) { r.RegisterProvider("ses", providerFromConfig) }
