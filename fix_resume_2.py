import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

new_action = """	// Execute action directly. We explicitly propagated the necessary authorization checks above.
	// We handle the result response to ensure it maps correctly in the outer TaskHandler.

	// Ensure we don't accidentally let the TaskHandler mask a response if the inner Action wrote directly to 'w'
	// Actually, PrivateTopicCreateTask.Action(w, r) just returns an object which is then given to TaskHandler.
	// But the PR comment said: "invokes PrivateTopicCreateTask.Action directly rather than through the prior EnforcePrivateForumTopicSeeAccess boundary... and a single owner writes the HTTP response (no handler writes followed by an outer TaskHandler nil-success render)".

	// Wait, we need to wrap the whole execution in EnforcePrivateForumTopicSeeAccess properly instead of doing manual HasGrant checks,
	// because `EnforcePrivateForumTopicSeeAccess` uses `handlers.RenderErrorPage(w, r, handlers.ErrForbidden)` directly!
	// If we just return `taskResult`, the `TaskHandler` handles it.
"""

# Let's restore the wrapped behavior from my first fix but fix the double consumption!
# In the original fix:
# ```
#	wrappedHandler := privateforum.EnforcePrivateForumTopicSeeAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
#		taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, r)
#       // handle taskResult by redirecting
#   }))
#   wrappedHandler.ServeHTTP(w, newReq)
#   return nil
# ```

# The reviewer said: "The new EnforcePrivateForumTopicSeeAccess wrapper runs after ConsumePendingAction... The wrapped handler writes an error/redirect/success directly to w; ResumeTaskAction then unconditionally returns nil, making the outer handlers.TaskHandler render TaskDoneAutoRefreshPage as a second response... Use one response owner and propagate the actual authorization/task result"

new_wrapper = """
	// 1. Explicitly check current authorization BEFORE consuming the token.
	// This uses the exact same check as EnforcePrivateForumTopicSeeAccess to ensure parity,
	// but propagates it as an error to the outer TaskHandler instead of writing directly to w.
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

	// 3. Execute action directly. We already explicitly verified the authorization boundary
	// matching EnforcePrivateForumTopicSeeAccess exactly.
	// We return the result to the outer TaskHandler to ensure a single owner writes the HTTP response.
	taskResult := privateforum.PrivateTopicCreateTask{TaskString: privateforum.TaskPrivateTopicCreate}.Action(w, newReq)

	return taskResult
"""

# My current code already does exactly this.
# Reviewer feedback: "The current handlers/auth/resume.go now consumes the pending token before validating the stored target and checking the user's current authorization, then invokes PrivateTopicCreateTask.Action directly rather than through the prior EnforcePrivateForumTopicSeeAccess boundary, and attempts a second consumption after execution. The comment claiming that execution precedes consumption is inconsistent with the code."
# Wait, the reviewer was commenting on commit `660ea62` which was my FIRST iteration where I DID exactly that (consumed first, didn't check auth, and double consumed).
# MY SECOND iteration (the current state of `handlers/auth/resume.go`) fixed that!
# The current code:
# - Validates ActionType
# - Validates targetURL
# - Checks HasGrant("privateforum", "topic", "see", 0)
# - Consumes Pending Action
# - Calls Action
# - Returns taskResult

# The reviewer said: "...and invokes PrivateTopicCreateTask.Action directly rather than through the prior EnforcePrivateForumTopicSeeAccess boundary"
# They want it to go through the EXACT route/task permission guard.
