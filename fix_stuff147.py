import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Make the exact diff change requested for `loginUserFunc`
content = content.replace(
"""	respPost, err := client.Do(reqPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respPost.StatusCode, "Login POST must succeed (200 OK after redirect)")
	bodyPost, _ := io.ReadAll(respPost.Body)
	require.NotContains(t, string(bodyPost), "Invalid credentials", "Login must succeed without invalid credentials error")
	respPost.Body.Close()

	// Prove authentication by fetching a protected endpoint (e.g. /usr)
	reqCheck, _ := http.NewRequest("GET", serverURL+"/usr", nil)
	respCheck, err := client.Do(reqCheck)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respCheck.StatusCode, "Must be authenticated and able to access /usr")
	respCheck.Body.Close()""",
"""	oldRedirect := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respPost, err := client.Do(reqPost)
	require.NoError(t, err)
	bodyPost, err := io.ReadAll(respPost.Body)
	require.NoError(t, err)
	respPost.Body.Close()
	require.Equal(t, http.StatusSeeOther, respPost.StatusCode, "successful login must redirect")
	require.NotContains(t, string(bodyPost), "Invalid username or password")
	client.CheckRedirect = oldRedirect

	// Prove that the same cookie jar now represents an authenticated session
	// without creating another pending-action nonce.
	reqVerify, err := http.NewRequest(http.MethodGet, serverURL+"/usr/logout", nil)
	require.NoError(t, err)
	respVerify, err := client.Do(reqVerify)
	require.NoError(t, err)
	defer respVerify.Body.Close()
	require.Equal(t, http.StatusOK, respVerify.StatusCode, "login must establish an authenticated session")"""
)

# And make sure I have `crypto/sha256`, `crypto/tls` etc imported!
# Wait, I might already have `crypto/tls` imported. If not, let's fix imports!
content = content.replace(
"""import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)""",
"""import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)"""
)

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)
