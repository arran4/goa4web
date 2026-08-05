//go:build ses

package ses

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	a4config "github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Provider wraps the AWS SES client.
type Provider struct {
	Client *ses.Client
	From   string
}

// Built indicates whether the SES provider is compiled in.
const Built = true

func (s Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	input := &ses.SendRawEmailInput{
		Destinations: []string{to.Address},
		Source:       aws.String(s.From),
		RawMessage:   &types.RawMessage{Data: rawEmailMessage},
	}
	_, err := s.Client.SendRawEmail(ctx, input)
	return err
}

func (s Provider) TestConfig(ctx context.Context) error {
	_, err := s.Client.GetSendQuota(ctx, &ses.GetSendQuotaInput{})
	if err != nil {
		return fmt.Errorf("failed to get send quota: %w", err)
	}
	fmt.Println("SES provider is configured correctly")
	return nil
}

func providerFromConfig(cfg *a4config.RuntimeConfig) (email.Provider, error) {
	opts := []func(*config.LoadOptions) error{}
	if region := cfg.EmailAWSRegion; region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("email disabled: cannot load AWS config: %w", err)
	}

	// Try retrieving credentials to see if they're available
	_, err = awsCfg.Credentials.Retrieve(context.Background())
	if err != nil {
		return nil, fmt.Errorf("email disabled: no AWS credentials: %w", err)
	}

	return Provider{Client: ses.NewFromConfig(awsCfg), From: cfg.EmailFrom}, nil
}

// Register registers the SES provider factory.
func Register(r *email.Registry) { r.RegisterProvider("ses", providerFromConfig) }
