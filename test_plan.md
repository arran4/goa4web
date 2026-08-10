1. **Define new SQL queries for retrieving grants**
   - In `internal/db/queries-private_forum.sql`, add `AdminListAllPrivateForumGrants` to fetch all grants for `privateforum` and `privateforum_thread` sections. Run `sqlc generate`.

2. **Add a new maintenance task and page for inconsistencies**
   - In `handlers/admin/tasks.go`, define `TaskCheckPrivateForumGrants`.
   - In `handlers/admin/admin_maintenance_privateforum_check_task.go`, implement `CheckPrivateForumGrantsTask`. It should present inconsistencies and fix checked ones.
   - In `handlers/admin/routes.go`, register the task with the `maintenance` endpoint.

3. **Modify maintenance page to link the new task**
   - In `handlers/admin/maintenancePage.go`, provide the `CheckTaskName` to the template.
   - In `core/templates/site/domains/admin/maintenancePage.gohtml`, add a button to submit the `TaskCheckPrivateForumGrants` task.

4. **Create a template to preview/fix inconsistencies**
   - Create `core/templates/site/domains/admin/maintenancePrivateForumCheckPreviewPage.gohtml`. List all found inconsistencies as a form with checkboxes.

5. **Implement the consistency check logic**
   - In `core/common/privateforum_check.go`, implement `CheckAndFixPrivateForumInconsistencies` that checks:
     - Rule 1: No "anyone" access (no role, no user set in grant).
     - Rule 2: Must specify an `item_id`.
     - Rule 3: Missing thread access when user has topic access. Since fixing this would involve creating missing grants (which would not have grant IDs to check), we will create "dummy" inconsistencies with `GrantID: 0` that insert missing grants instead of deleting.

6. **Refine Rule 3 (Fixing missing thread access)**
   - In `core/common/privateforum_check.go`, add logic to check if a user has access to a topic but is missing access to one of its threads.
   - When returning these as inconsistencies, use a negative `GrantID` or another mechanism so that we can distinguish it from a deletion. When a user selects to fix it, it will execute a `SystemCreateGrant` for that thread.

7. **Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.**

8. **Submit changes**
