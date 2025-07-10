//go:build s3

package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/upload"
)

// Built indicates whether the S3 provider is compiled in.
const Built = true

type api interface {
	HeadBucket(*awsS3.HeadBucketInput) (*awsS3.HeadBucketOutput, error)
	PutObject(*awsS3.PutObjectInput) (*awsS3.PutObjectOutput, error)
	DeleteObject(*awsS3.DeleteObjectInput) (*awsS3.DeleteObjectOutput, error)
	GetObject(*awsS3.GetObjectInput) (*awsS3.GetObjectOutput, error)
}

type Provider struct {
	Client api
	Bucket string
	Prefix string
}

func providerFromConfig(cfg config.RuntimeConfig) upload.Provider {
	raw := cfg.ImageUploadS3URL
	if raw == "" {
		raw = cfg.ImageUploadDir
	}
	b, p, err := parseURL(raw)
	if err != nil {
		return nil
	}
	c, err := newClient(cfg.EmailAWSRegion)
	if err != nil {
		return nil
	}
	return Provider{Client: c, Bucket: b, Prefix: p}
}

func Register() { upload.RegisterProvider("s3", providerFromConfig) }

var newClient = func(region string) (api, error) {
	cfg := aws.NewConfig()
	if region != "" {
		cfg = cfg.WithRegion(region)
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	if _, err := sess.Config.Credentials.Get(); err != nil {
		return nil, err
	}
	return awsS3.New(sess), nil
}

func parseURL(raw string) (bucket, prefix string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	if u.Scheme != "s3" || u.Host == "" {
		return "", "", fmt.Errorf("invalid s3 url")
	}
	bucket = u.Host
	prefix = strings.TrimPrefix(u.Path, "/")
	return bucket, prefix, nil
}

func (p Provider) Check(ctx context.Context) error {
	if _, err := p.Client.HeadBucket(&awsS3.HeadBucketInput{Bucket: aws.String(p.Bucket)}); err != nil {
		return err
	}
	key := path.Join(p.Prefix, ".check")
	if _, err := p.Client.PutObject(&awsS3.PutObjectInput{Bucket: aws.String(p.Bucket), Key: aws.String(key), Body: bytes.NewReader([]byte("ok"))}); err != nil {
		return err
	}
	if _, err := p.Client.DeleteObject(&awsS3.DeleteObjectInput{Bucket: aws.String(p.Bucket), Key: aws.String(key)}); err != nil {
		return err
	}
	return nil
}

func (p Provider) Write(ctx context.Context, name string, data []byte) error {
	key := path.Join(p.Prefix, name)
	_, err := p.Client.PutObject(&awsS3.PutObjectInput{Bucket: aws.String(p.Bucket), Key: aws.String(key), Body: bytes.NewReader(data)})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

func (p Provider) Read(ctx context.Context, name string) ([]byte, error) {
	key := path.Join(p.Prefix, name)
	out, err := p.Client.GetObject(&awsS3.GetObjectInput{Bucket: aws.String(p.Bucket), Key: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
