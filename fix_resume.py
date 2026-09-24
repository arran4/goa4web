import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

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

	// 1. Explicitly check current authorization BEFORE consuming the token
	// This ensures we do not burn the token if the user lacks authorization right now
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		return handlers.ErrForbidden
	}

	// 2. Consume atomically AFTER authorization, but BEFORE execution
	// This ensures exactly-once execution (at-most-once under crashes)
	rows, err := cd.Queries().ConsumePendingAction(r.Context(), tokenHashHex)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	newReq := r.Clone(r.Context())
	newReq.URL = targetURL
	newReq.Method = http.MethodPost
	newReq.PostForm = storageMap.Form

	// Execute action directly. We explicitly propagated the necessary authorization checks above.
	// We handle the result response to ensure it maps correctly in the outer TaskHandler.
	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
}"""

content = re.sub(r"func ResumeTaskAction\(w http.ResponseWriter, r \*http.Request\) any \{.*?\n\}\n\ntype ResumeTask", new_action + "\n\ntype ResumeTask", content, flags=re.DOTALL)

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)
