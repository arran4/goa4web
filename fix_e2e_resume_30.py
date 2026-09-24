import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Our explicit resume test for Bob got a 404. Why?
# "token := r.PostFormValue("token")"
# "action, err := cd.Queries().GetPendingAction(r.Context(), tokenHashHex)"
# Why would GetPendingAction return 404 for Bob's token after relogging in?
# Because bob got a NEW browser_id when logging out and logging in?
# Wait! "Bob explicitly logs out cleanly using the true flow" - POST /usr/logout doesn't clear a4w_bid! But maybe clientC's cookie jar handles it? It shouldn't clear it.
# Actually, Bob's token might be 404 because the interceptor DIDN'T save it properly!
# "require.NotEmpty(t, resumeTokenC)"
# Let's check `interceptStalePost`. It requires `browserID := core.GetBrowserID(w, r)`.
# Ah! When Bob logs out, his `a4w_bid` cookie IS preserved.
# But wait, when Bob POSTs to `/resume`, the CSRF validation might fail if `gorilla.csrf.Token` isn't correct.
# Wait, if CSRF fails, the handler is not called, and the router renders `RenderNotFoundOrLogin` which usually returns 404 if not logged in. Wait, Bob IS logged in, so it returns 404!
# Let's verify `loginCsrfC` was properly set.

content = content.replace('reqResumeAuthCheck.Header.Set("Content-Type", "application/x-www-form-urlencoded")', 'reqResumeAuthCheck.Header.Set("Content-Type", "application/x-www-form-urlencoded")\n\tclientC.CheckRedirect = nil')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)
