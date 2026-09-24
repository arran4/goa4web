import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Why does Bob get a 404 instead of 403?
# The handler ResumeTaskAction returns handlers.ErrNotFound when action does not match.
# Wait, let's look at ResumeTaskAction logic:
# `action, err := cd.Queries().GetPendingAction(r.Context(), tokenHashHex)`
# `if action.Uid != cd.UserID || action.BrowserID != browserID { return handlers.ErrForbidden }`
# Is the Uid matching? Bob is logging in as bob. Is Bob's UID = 2? Wait!
# In the testdata `scenarios/100-private-forum`, the users are "admin" (ID=1) and "bob" (ID=maybe not 2??).
# Wait, user_id=2 might be someone else, or Bob might be ID=2.
# Wait, Bob's browser_id when he logs out might be lost if `clientC.Jar` is wiped?
# We used `loginUserFunc` which doesn't clear `clientC.Jar`, it just gets a new session.
# BUT `core.GetBrowserID` uses a separate unauthenticated cookie "a4w_bid".
# Does `loginUserFunc` accidentally clear cookies? No, `cookiejar` persists `a4w_bid`.
# What about `GetPendingAction(..., tokenHashHex)`?
# Let's see how `resumeTokenC` is parsed: `strings.TrimPrefix(locC, "/login?resume=")`
# Oh! The interceptor redirects to `/login?back=/resume?token=...` !!
# `backURL := "/resume?token=" + resumeToken`
# `newVals.Set("back", backURL)`
# `target := "/login?" + newVals.Encode()`
# So `Location` is `/login?back=%2Fresume%3Ftoken%3D...`!
# `strings.TrimPrefix` on `/login?resume=` won't work!
# Wait! In our original successful test:
# `resumeToken := extractResumeToken(location)`
# Let's use THAT function! `extractResumeToken` !

content = content.replace('resumeTokenC := strings.TrimPrefix(locC, "/login?resume=")', 'resumeTokenC := extractResumeToken(locC)')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)
