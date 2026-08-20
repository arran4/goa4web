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
			name:     "URL with tracking param first",
			input:    "https://example.com/item?utm_source=x&id=1&category=books",
			expected: "https://example.com/item?category=books&id=1",
		},
		{
			name:     "URL with tracking param middle",
			input:    "https://example.com/item?id=1&utm_source=x&category=books",
			expected: "https://example.com/item?category=books&id=1",
		},
		{
			name:     "URL with tracking param last",
			input:    "https://example.com/item?id=1&category=books&utm_source=x",
			expected: "https://example.com/item?category=books&id=1",
		},
		{
			name:     "URL with multiple tracking params at start",
			input:    "https://example.com/item?utm_source=x&utm_medium=y&id=1&category=books",
			expected: "https://example.com/item?category=books&id=1",
		},
		{
			name:     "URL with fbclid, gclid, gbraid, wbraid, mc_cid, mc_eid, igshid, msclkid, twclid, yclid, click_id, clickid, _hsenc, _hsmi, mkt_tok",
			input:    "https://example.com/deal?fbclid=1&gclid=2&gbraid=3&wbraid=4&mc_cid=5&mc_eid=6&igshid=7&msclkid=8&twclid=9&yclid=10&click_id=11&clickid=12&_hsenc=13&_hsmi=14&mkt_tok=15&product_id=99",
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
