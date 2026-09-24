import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

# We need to change the resume task action to use cd.HasGrant BEFORE consumption and throw a true 403!
# And it should use `cd.HasGrant("privateforum", "topic", "create", 0)` too since the action checks for it!
# Wait, why was `TaskHandler` throwing 500 when it returned `handlers.ErrForbidden`?
# In `handlers/taskhandler.go`:
# `case error:`
# `var ue interface { error, UserErrorMessage() string }`
# `if errors.As(result, &ue) { ... TaskErrorAcknowledgementPage(w, r) return }`
# `handlers.ErrForbidden` IS an error. It does NOT have `UserErrorMessage()`.
# So it falls through to `RenderErrorPage(w, r, result)`.
# `RenderErrorPage` does:
# `var he *HTTPError`
# `if errors.As(err, &he) { status = he.Status }`
# BUT `handlers.ErrForbidden` is `*HTTPError` and has `Status = 403`.
# So `RenderErrorPage` sets `status = 403` and calls `w.WriteHeader(403)`.
# Then it renders the template `TaskErrorAcknowledgementPageTmpl`.
# If `cd.ExecuteSiteTemplate` fails, it writes 500.
# WHY WOULD IT FAIL?
# Wait! In the logs: `task action: create private topic permission denied`
# This means `ResumeTaskAction` returned `fmt.Errorf("create private topic permission denied")`!
# Ah! It didn't return `handlers.ErrForbidden`. It returned an error from `PrivateTopicCreateTask.Action`!
# Because the `see` grant passed, so `ResumeTaskAction` did NOT return `handlers.ErrForbidden`. It proceeded to consumption, and then the task action failed on the `create` grant check!
# THIS means we consumed the token, and then got a 500 from the task failing!
# So to fix this, we MUST check `cd.HasGrant("privateforum", "topic", "create", 0)` BEFORE consumption!

new_action = """func ResumeTaskAction(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	token := r.PostFormValue("token")
	if token == "" {
		return handlers.ErrNotFound
	}
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	browserID := core.GetBrowserID(w, r)

	action, err := cd.Queries().GetPendingAction(r.Context(), tokenHashHex)
	if err != nil {
		return handlers.ErrNotFound
	}

	if action.Uid != cd.UserID || action.BrowserID != browserID {
		return handlers.ErrForbidden
	}

	if action.ActionType != string(privateforum.TaskPrivateTopicCreate) {
		return handlers.ErrForbidden
	}

	var storageMap struct {
		Form url.Values `json:"form"`
		URL  string     `json:"url"`
	}
	if err := json.Unmarshal([]byte(action.FormData), &storageMap); err != nil {
		return fmt.Errorf("invalid form data")
	}

	targetURL, err := url.Parse(storageMap.URL)
	if err != nil || targetURL.IsAbs() || targetURL.Host != "" || targetURL.Path != "/private/topic/new" {
		return handlers.ErrForbidden
	}

	// 1. Explicitly check current authorization BEFORE consuming the token.
	// We want to preserve the token if the user is legitimate but simply unauthorized right now.
	if !cd.HasGrant("privateforum", "topic", "see", 0) || !cd.HasGrant("privateforum", "topic", "create", 0) {
		return handlers.ErrForbidden
	}

	// 2. Consume atomically AFTER authorization checks to ensure exactly-once semantics.
	// Since we don't have global explicit multi-statement transactions in the app's framework
	// for arbitrary actions, we at least prevent duplicate submission concurrency natively via
	// the Consume SQL query, and error cleanly on partial failure without duplicate topics.
	rows, err := cd.Queries().ConsumePendingAction(r.Context(), tokenHashHex)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	newReq := r.Clone(r.Context())
	newReq.URL = targetURL
	newReq.Method = http.MethodPost
	newReq.PostForm = storageMap.Form

	// Execute action directly, returning its result properly to the outer TaskHandler
	// since we already explicitly checked the authorization constraint above.
	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
}"""

content = re.sub(r"func ResumeTaskAction\(w http.ResponseWriter, r \*http.Request\) any \{.*?\n\}\n\ntype ResumeTask", new_action + "\n\ntype ResumeTask", content, flags=re.DOTALL)

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)
