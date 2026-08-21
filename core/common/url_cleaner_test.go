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
			name:     "Tracking param first - order preserved",
			input:    "https://example.com/item?utm_source=x&id=1&category=books",
			expected: "https://example.com/item?id=1&category=books",
		},
		{
			name:     "Tracking param middle - order preserved",
			input:    "https://example.com/item?id=1&utm_source=x&category=books",
			expected: "https://example.com/item?id=1&category=books",
		},
		{
			name:     "Tracking param last - order preserved",
			input:    "https://example.com/item?id=1&category=books&utm_source=x",
			expected: "https://example.com/item?id=1&category=books",
		},
		{
			name:     "Multiple tracking params at start - order preserved",
			input:    "https://example.com/item?utm_source=x&utm_medium=y&id=1&category=books",
			expected: "https://example.com/item?id=1&category=books",
		},
		{
			name:     "Representation: space encoding %20 preserved without re-encoding to +",
			input:    "https://example.com/search?q=hello%20world&utm_source=x",
			expected: "https://example.com/search?q=hello%20world",
		},
		{
			name:     "Representation: escaped slash %2F preserved",
			input:    "https://example.com/search?q=a%2Fb&utm_source=x",
			expected: "https://example.com/search?q=a%2Fb",
		},
		{
			name:     "Representation: repeated parameters and their order preserved",
			input:    "https://example.com/search?a=1&a=2&utm_source=x",
			expected: "https://example.com/search?a=1&a=2",
		},
		{
			name:     "Representation: empty raw query components preserved between params",
			input:    "https://example.com/search?a=1&&utm_source=x&b=2",
			expected: "https://example.com/search?a=1&&b=2",
		},
		{
			name:     "Representation: leading empty raw query component preserved",
			input:    "https://example.com/search?&a=1&utm_source=x",
			expected: "https://example.com/search?&a=1",
		},
		{
			name:     "Representation: trailing empty raw query component preserved",
			input:    "https://example.com/search?a=1&utm_source=x&",
			expected: "https://example.com/search?a=1&",
		},
		{
			name:     "Representation: multiple consecutive empty raw query components preserved",
			input:    "https://example.com/search?a=1&&utm_source=x&&b=2",
			expected: "https://example.com/search?a=1&&&b=2",
		},
		{
			name:     "Representation: percent-encoded key not treated as tracking key (conservative raw matching)",
			input:    "https://example.com/search?%75tm_source=x&id=1",
			expected: "https://example.com/search?%75tm_source=x&id=1",
		},
		{
			name:     "UTM prefix matching with hyphen in key",
			input:    "https://example.com/item?utm_campaign-name=x&id=1",
			expected: "https://example.com/item?id=1",
		},
		{
			name:     "UTM prefix matching with dot in key",
			input:    "https://example.com/item?utm_custom.value=x&id=1",
			expected: "https://example.com/item?id=1",
		},
		{
			name:     "UTM uppercase prefix matching with hyphen in key",
			input:    "https://example.com/item?UTM_custom-value=x&id=1",
			expected: "https://example.com/item?id=1",
		},
		{
			name:     "Non-UTM prefix key utm.foo preserved",
			input:    "https://example.com/item?utm.foo=x&id=1",
			expected: "https://example.com/item?utm.foo=x&id=1",
		},
		{
			name:     "Preserve exact-key prefixes like gclid_extra when tracking param is also removed",
			input:    "https://example.com/?id=1&gclid_extra=keep&utm_source=x",
			expected: "https://example.com/?id=1&gclid_extra=keep",
		},
		{
			name:     "Preserve fbclid_extra when tracking param is also removed",
			input:    "https://example.com/?id=1&fbclid_extra=val&utm_medium=mail",
			expected: "https://example.com/?id=1&fbclid_extra=val",
		},
		{
			name:     "Preserve clickidfoo when tracking param is also removed",
			input:    "https://example.com/?id=1&clickidfoo=123&gclid=real",
			expected: "https://example.com/?id=1&clickidfoo=123",
		},
		{
			name:     "Nested ? in functional query value is preserved intact",
			input:    "https://example.com/?redirect=https://target.test/path?utm_source=x",
			expected: "https://example.com/?redirect=https://target.test/path?utm_source=x",
		},
		{
			name:     "Nested ? in functional query value preserved while outer tracker is removed",
			input:    "https://example.com/?redirect=https://target.test/path?utm_source=x&gclid=real",
			expected: "https://example.com/?redirect=https://target.test/path?utm_source=x",
		},
		{
			name:     "? in fragment is preserved without outer query",
			input:    "https://example.com/path#section?utm_source=x",
			expected: "https://example.com/path#section?utm_source=x",
		},
		{
			name:     "? in fragment is preserved while outer tracker is removed",
			input:    "https://example.com/path?id=1&utm_source=x#section?gclid=fragment",
			expected: "https://example.com/path?id=1#section?gclid=fragment",
		},
		{
			name:     "&tracking in fragment is preserved without being cleaned",
			input:    "https://example.com/path?id=1#section&utm_source=x",
			expected: "https://example.com/path?id=1#section&utm_source=x",
		},
		{
			name:     "Signature-looking fragment does not prevent outer tracker removal",
			input:    "https://example.com/?utm_source=x#section&signature=not-query",
			expected: "https://example.com/#section&signature=not-query",
		},
		{
			name:     "Empty host https://?utm_source=x is left untouched",
			input:    "https://?utm_source=x",
			expected: "https://?utm_source=x",
		},
		{
			name:     "Empty host https:///path?utm_source=x is left untouched",
			input:    "https:///path?utm_source=x",
			expected: "https:///path?utm_source=x",
		},
		{
			name:     "Representation: blank value empty= preserved",
			input:    "https://example.com/search?empty=&utm_source=x",
			expected: "https://example.com/search?empty=",
		},
		{
			name:     "Representation: fragment with retained params preserved",
			input:    "https://example.com/search?id=1&utm_source=x#fragment",
			expected: "https://example.com/search?id=1#fragment",
		},
		{
			name:     "Representation: fragment with only tracking params preserved",
			input:    "https://example.com/search?utm_source=x#fragment",
			expected: "https://example.com/search#fragment",
		},
		{
			name:     "URL with all known tracking key families",
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
			name:     "Representation: uppercase scheme case preserved",
			input:    "HTTPS://example.com/item?id=1&utm_source=x",
			expected: "HTTPS://example.com/item?id=1",
		},
		{
			name:     "Representation: uppercase scheme with only tracking params",
			input:    "HTTPS://example.com/item?utm_source=x",
			expected: "HTTPS://example.com/item",
		},
		{
			name:     "Representation: leading empty component left as bare question mark",
			input:    "https://example.com/search?&utm_source=x",
			expected: "https://example.com/search?",
		},
		{
			name:     "Representation: trailing empty component left as bare question mark",
			input:    "https://example.com/search?utm_source=x&",
			expected: "https://example.com/search?",
		},
		{
			name:     "Representation: leading and trailing empty components left as ?&",
			input:    "https://example.com/search?&utm_source=x&",
			expected: "https://example.com/search?&",
		},
		{
			name:     "Representation: multiple trailing empty components left as ?&",
			input:    "https://example.com/search?utm_source=x&&",
			expected: "https://example.com/search?&",
		},
		{
			name:     "Representation: unusual host, userinfo, port, path escaping, and fragment preserved",
			input:    "https://User:Pass@EXAMPLE.com:8080/path/to/page%201?id=42&utm_source=x#section-1",
			expected: "https://User:Pass@EXAMPLE.com:8080/path/to/page%201?id=42#section-1",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Non-http scheme: mailto left untouched",
			input:    "mailto:user@example.com?utm_source=x",
			expected: "mailto:user@example.com?utm_source=x",
		},
		{
			name:     "Non-http scheme: ftp left untouched",
			input:    "ftp://example.com/file?utm_source=x",
			expected: "ftp://example.com/file?utm_source=x",
		},
		{
			name:     "Scheme-relative URL left untouched",
			input:    "//example.com/file?utm_source=x",
			expected: "//example.com/file?utm_source=x",
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
