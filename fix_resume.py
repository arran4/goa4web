import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

# Replace the body of ResumeTaskAction
new_body = """func ResumeTaskAction(w http.ResponseWriter, r *http.Request) any {
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

	// Consume atomically BEFORE execution to ensure exactly-once semantics.
	// If the database fails or validation fails below, the action is burned,
	// prioritizing duplicate-prevention over automatic retry.
	rows, err := cd.Queries().ConsumePendingAction(r.Context(), tokenHashHex)
	if err != nil || rows == 0 {
		return handlers.ErrNotFound
	}

	newReq := r.Clone(r.Context())
	newReq.URL = targetURL
	newReq.Method = http.MethodPost
	newReq.PostForm = storageMap.Form

	// Execute action wrapped in the original authorization boundary
	wrappedHandler := privateforum.EnforcePrivateForumTopicSeeAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, r)
		if err, ok := taskResult.(error); ok {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		if red, ok := taskResult.(core.RedirectResponse); ok {
			http.Redirect(w, r, red.RedirectPath, http.StatusSeeOther)
			return
		}
	}))

	// Create a response recorder to capture the result
	wrappedHandler.ServeHTTP(w, newReq)

	return nil
}"""

content = re.sub(r"func ResumeTaskAction\(w http.ResponseWriter, r \*http.Request\) any \{.*?\n\}\n\ntype ResumeTask", new_body + "\n\ntype ResumeTask", content, flags=re.DOTALL)

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)
