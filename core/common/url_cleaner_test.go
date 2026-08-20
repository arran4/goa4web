package common

import (
	"testing"
)

func TestCanonicalizeExternalURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain URL without query",
			input:    "https://example.com/article/123",
			expected: "https://example.com/article/123",
		},
		{
			name:     "URL with clean functional params",
			input:    "https://example.com/search?q=golang&page=2",
			expected: "https://example.com/search?q=golang&page=2",
		},
		{
			name:     "URL with utm parameters only",
			input:    "https://example.com/article?utm_source=twitter&utm_medium=social&utm_campaign=summer_sale",
			expected: "https://example.com/article",
		},
		{
			name:     "URL with mixed functional and utm parameters",
			input:    "https://example.com/item?id=42&utm_source=newsletter&category=books",
			expected: "https://example.com/item?category=books&id=42",
		},
		{
			name:     "URL with fbclid, gclid, yclid, click_id",
			input:    "https://example.com/deal?fbclid=12345&gclid=67890&yclid=abcde&click_id=xyz&product_id=99",
			expected: "https://example.com/deal?product_id=99",
		},
		{
			name:     "Preserve Bloomberg accessToken and resource",
			input:    "https://www.bloomberg.com/news/articles/2026-01-01/sample?accessToken=secret_token_123&resource=finance",
			expected: "https://www.bloomberg.com/news/articles/2026-01-01/sample?accessToken=secret_token_123&resource=finance",
		},
		{
			name:     "Preserve X-Amz-Security-Token without signature",
			input:    "https://s3.amazonaws.com/bucket/key?X-Amz-Security-Token=token123&versionId=v1",
			expected: "https://s3.amazonaws.com/bucket/key?X-Amz-Security-Token=token123&versionId=v1",
		},
		{
			name:     "Preserve unknown parameters",
			input:    "https://example.com/test?custom_flag=true&my_app_param=active&utm_source=bad",
			expected: "https://example.com/test?custom_flag=true&my_app_param=active",
		},
		{
			name:     "AWS presigned URL with X-Amz-Signature is kept completely untouched",
			input:    "https://mybucket.s3.amazonaws.com/file.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20260820%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260820T000000Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=abcdef1234567890&utm_source=presigned",
			expected: "https://mybucket.s3.amazonaws.com/file.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20260820%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260820T000000Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=abcdef1234567890&utm_source=presigned",
		},
		{
			name:     "CloudFront signed URL with Signature parameter is kept untouched",
			input:    "https://d111111abcdef8.cloudfront.net/video.mp4?Key-Pair-Id=K2JCJMDEHXQW5F&Signature=fFakeSignatureString12345&Expires=1750000000&utm_medium=video",
			expected: "https://d111111abcdef8.cloudfront.net/video.mp4?Key-Pair-Id=K2JCJMDEHXQW5F&Signature=fFakeSignatureString12345&Expires=1750000000&utm_medium=video",
		},
		{
			name:     "URL with generic sig / hmac signature is kept untouched",
			input:    "https://api.example.com/download?file=data.zip&sig=d8e8fca2dc0f896fd7cb4cb0031ba249&utm_campaign=dl",
			expected: "https://api.example.com/download?file=data.zip&sig=d8e8fca2dc0f896fd7cb4cb0031ba249&utm_campaign=dl",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid or non-http URL",
			input:    "mailto:test@example.com?subject=Hello",
			expected: "mailto:test@example.com?subject=Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanonicalizeExternalURL(tt.input)
			if got != tt.expected {
				t.Errorf("CanonicalizeExternalURL(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
