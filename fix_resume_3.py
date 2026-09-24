import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

# We need to use `EnforcePrivateForumTopicSeeAccess` as requested: "rather than through the prior EnforcePrivateForumTopicSeeAccess boundary... validate payload and current grants before performing the write, with no route that bypasses the original authorization boundary."
# We can do this by creating a wrapper that DOES the consumption inside the authenticated boundary!
# But wait, we want to return the result to TaskHandler!
# Wait! ResumeTaskAction IS a task action.
# So we can't easily use an `http.Handler` middleware without dealing with response writers.
# But `privateforum.EnforcePrivateForumTopicSeeAccess` is just an `http.Handler` middleware.

# To satisfy "validate payload and current grants before performing the write, with no route that bypasses the original authorization boundary. Claim/consume once, atomically"
# Let's write the response properly using a recorder OR just let TaskHandler do it.
# Actually, the best way to enforce it is to route the request through the actual Router!
# But then we get infinite loops or double task execution?
# "Restore strict checks for owner UID, same browser, expected ActionType, and an exact safe relative /private/topic/new target; validate payload and current grants before performing the write, with no route that bypasses the original authorization boundary."

# Let's use `EnforcePrivateForumTopicSeeAccess` manually but correctly intercept the error.

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

	// 1. Explicitly invoke the exact original authorization boundary
	var authErr error
	authBoundary := privateforum.EnforcePrivateForumTopicSeeAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inside the boundary, authorization succeeded.
		authErr = nil
	}))

	// Create a dummy writer to capture any early rejection without writing to the real client yet
	rw := httptest.NewRecorder()
	authBoundary.ServeHTTP(rw, r)

	if rw.Code == http.StatusForbidden {
		return handlers.ErrForbidden
	}
	if rw.Code != http.StatusOK {
		// If the middleware rejected it with some other code, propagate an error
		return fmt.Errorf("authorization rejected with code %d", rw.Code)
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

	// Execute action directly, returning its result properly to the outer TaskHandler.
	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
}"""

content = re.sub(r"func ResumeTaskAction\(w http.ResponseWriter, r \*http.Request\) any \{.*?\n\}\n\ntype ResumeTask", new_action + "\n\ntype ResumeTask", content, flags=re.DOTALL)

if '"net/http/httptest"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"net/http/httptest"')

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)
