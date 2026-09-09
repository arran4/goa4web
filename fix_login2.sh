cat << 'HTML_EOF' > core/templates/site/pages/auth/loginPage.gohtml
{{ template "head" $ }}
    {{- if cd.CurrentNotice }}
    <p>{{ cd.CurrentNotice }}</p>
    {{- end }}
    {{- if cd.CurrentError }}
    <p class="text-error">{{ cd.CurrentError }}</p>
    {{- end }}
    Login:<br>
    <form method="post" action="/login">
        {{ csrfField }}
        {{- if $.Back }}
        <input type="hidden" name="back" value="{{ $.Back }}">
        {{- end }}
        {{- if $.Code }}
        <input type="hidden" name="code" value="{{ $.Code }}">
        {{- end }}
        Username: <input name="username"><br>
        Password: <input name="password" type="password"><br>
        <input type="submit" name="task" value="Login">
        <br><a href="/forgot">Forgot password?</a>
    </form>
    <br>
    <hr>
    {{ if cd.WebAuthn }}<h3>Login with Passkey</h3>
    <form id="passkey-login-form">
        {{ csrfField }}
        Username: <input type="text" id="passkey-username" name="username" required><br>
        <button type="submit">Login with Passkey</button>
    </form>

    <script src="{{ assetHash "/static/passkeys.js" }}"></script>
{{ end }}
{{ template "tail" $ }}
HTML_EOF

cat << 'LP_EOF' > handlers/auth/loginPage.go
package auth

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

type loginFormHandler struct{ msg string }

func (l loginFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handlers.DisableCaching(w)
	renderLoginForm(w, r, l.msg, "")
}

var _ http.Handler = (*loginFormHandler)(nil)

const (
	TaskDoneAutoRefreshPageTmpl tasks.Template = "pages/misc/taskDoneAutoRefreshPage.gohtml"
)

func renderLoginForm(w http.ResponseWriter, r *http.Request, errMsg, noticeMsg string) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.SetCurrentError(errMsg)
	cd.SetCurrentNotice(noticeMsg)
	type Data struct {
		Code    string
		Back    string
	}
	handlers.SetPageTitle(r, "Login")
	backURL, _ := cd.SanitizeBackURL(r, r.FormValue("back"))
	data := Data{
		Code:    r.FormValue("code"),
		Back:    backURL,
	}
	_ = LoginPageTmpl.Handle(w, r, data)
}

const LoginPageTmpl tasks.Template = "pages/auth/loginPage.gohtml"
LP_EOF
