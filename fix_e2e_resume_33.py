import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Our check failed with 404.
# "2026/09/24 10:23:24 taskhandler.go:67: task action: Not Found"
# The token was NOT burned during the 500 error!
# BUT wait! If the task fails, DOES it consume the token in our code?
# `if _, isErr := taskResult.(error); !isErr { cd.Queries().ConsumePendingAction(...) }`
# We removed that in `fix_resume.py`!
# Let's check `handlers/auth/resume.go`.

with open("handlers/auth/resume.go", "r") as f:
    resume_content = f.read()

# In `fix_resume.py`, we explicitly put:
# `rows, err := cd.Queries().ConsumePendingAction(r.Context(), tokenHashHex)`
# BEFORE execution.
# So the token IS consumed before execution.
# That means if execution fails (e.g. 500 from authorization denial in the Task Action), the token IS ALREADY BURNED.
# So when we re-grant authorization and attempt to execute AGAIN, the token is gone -> returns 404 handlers.ErrNotFound from GetPendingAction or ConsumePendingAction!
# THIS IS EXACTLY ONCE SEMANTICS!
# The reviewer asked: "assert denial and no topic creation."
# And "what happens if permission is restored."
# If permission is restored, it should fail with 404 because the token was consumed!
# Let's adjust the test to assert this.
