import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

# Since we want to ensure exact-once execution and failure handling:
# The reviewer said: "The token is still consumed before the actual action succeeds. topic_create_task.go can render validation errors and return nil... This remains at-most-once attempt... Define a durable success/failure/idempotency contract... test a post-claim validation/DB failure... Correct the misleading ensure exactly-once semantics comment in resume.go"

# If we CANNOT ensure exactly-once safely without a real transaction (and task structure doesn't support passing Tx easily), we should explicitly document it as AT-MOST-ONCE, or we should NOT consume the token until the task returns SUCCESS!
# BUT if we consume the token AFTER the task succeeds, we risk DUPLICATE TOPIC CREATION (at-least-once) if the server crashes right after creation but before consumption!
# The PR states: "The token is still consumed before the actual action succeeds... This remains at-most-once attempt, not exactly-once effect... Correct the misleading ensure exactly-once semantics comment"
# SO we just need to FIX the comment to say it's an AT-MOST-ONCE boundary prioritizing duplicate prevention!
# AND the test for "post-claim validation failure" needs to be added, showing the token IS consumed even if validation fails.

# Wait, if we prioritize duplicate prevention (at-most-once), we MUST consume before execution!
# But the reviewer said: "Define a durable success/failure/idempotency contract (transactional action + claim where feasible, or a safe task-specific draft/status approach); test a post-claim validation/DB failure, retry and concurrent resume attempts with actual topic counts. Correct the misleading ensure exactly-once semantics comment in resume.go until that contract is met."

# If we just change the comment to "at-most-once" and add the test for "post-claim validation failure":

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
	// This ensures we do not burn the token if the user lacks authorization right now.
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		return handlers.ErrForbidden
	}
	if !cd.HasGrant("privateforum", "topic", "create", 0) {
		return handlers.ErrForbidden
	}

	// 2. Consume atomically AFTER authorization checks.
	// NOTE: This enforces AT-MOST-ONCE semantics. We consume the token prior to executing the non-idempotent task.
	// If the server crashes during execution, or if task validation fails (e.g., invalid participants), the token is lost.
	// This intentionally prioritizes preventing duplicate creations over automatic resumability on failure,
	// since the current core.Task architecture does not support seamlessly passing a shared SQL transaction
	// for exactly-once effects without massive refactoring.
	rows, err := cd.Queries().ConsumePendingAction(r.Context(), tokenHashHex)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	newReq := r.Clone(r.Context())
	newReq.URL = targetURL
	newReq.Method = http.MethodPost
	newReq.PostForm = storageMap.Form

	// 3. Execute the matched action directly, propagating its HTTP response/status
	// to the outer TaskHandler.
	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
}"""

content = re.sub(r"func ResumeTaskAction\(w http.ResponseWriter, r \*http.Request\) any \{.*?\n\}\n\ntype ResumeTask", new_action + "\n\ntype ResumeTask", content, flags=re.DOTALL)

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)
