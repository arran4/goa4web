//go:build s3

package s3

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/arran4/goa4web/runtimeconfig"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

type mockClient struct {
	headCalled bool
	putCalled  bool
	delCalled  bool
	getCalled  bool
	headErr    error
	putErr     error
	delErr     error
	getErr     error
	getData    []byte
}

func (m *mockClient) HeadBucket(*awsS3.HeadBucketInput) (*awsS3.HeadBucketOutput, error) {
	m.headCalled = true
	return &awsS3.HeadBucketOutput{}, m.headErr
}

func (m *mockClient) PutObject(*awsS3.PutObjectInput) (*awsS3.PutObjectOutput, error) {
	m.putCalled = true
	return &awsS3.PutObjectOutput{}, m.putErr
}

func (m *mockClient) DeleteObject(*awsS3.DeleteObjectInput) (*awsS3.DeleteObjectOutput, error) {
	m.delCalled = true
	return &awsS3.DeleteObjectOutput{}, m.delErr
}

func (m *mockClient) GetObject(*awsS3.GetObjectInput) (*awsS3.GetObjectOutput, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &awsS3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(m.getData))}, nil
}

func TestProviderCheckSuccess(t *testing.T) {
	orig := newClient
	mock := &mockClient{}
	newClient = func(string) (api, error) { return mock, nil }
	defer func() { newClient = orig }()

	p := providerFromConfig(runtimeconfig.RuntimeConfig{EmailAWSRegion: "us-east-1", ImageUploadS3URL: "s3://bucket/path"})
	if p == nil {
		t.Fatal("nil provider")
	}
	if err := p.Check(nil); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !mock.headCalled || !mock.putCalled || !mock.delCalled {
		t.Fatalf("expected calls")
	}
}

func TestProviderCheckWriteError(t *testing.T) {
	orig := newClient
	mock := &mockClient{putErr: fmt.Errorf("fail")}
	newClient = func(string) (api, error) { return mock, nil }
	defer func() { newClient = orig }()

	p := providerFromConfig(runtimeconfig.RuntimeConfig{ImageUploadS3URL: "s3://bucket/path"})
	if err := p.Check(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestProviderRead(t *testing.T) {
	orig := newClient
	mock := &mockClient{getData: []byte("hello")}
	newClient = func(string) (api, error) { return mock, nil }
	defer func() { newClient = orig }()

	p := providerFromConfig(runtimeconfig.RuntimeConfig{ImageUploadS3URL: "s3://bucket/path"})
	if p == nil {
		t.Fatal("nil provider")
	}
	data, err := p.Read(nil, "name")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "hello" || !mock.getCalled {
		t.Fatalf("unexpected %q %v", data, mock.getCalled)
	}
}
