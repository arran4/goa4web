package handlers

const (
	// Alphabet lists letters for search navigation.
	Alphabet = "abcdefghijklmnopqrstuvwxyz"

	// ExpectedSchemaVersion defines the required database schema version.
	// Bump this when adding a new migration.
<<<<<<< HEAD
	ExpectedSchemaVersion = 94
=======
	ExpectedSchemaVersion = 93
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)

	// CSRFField is the name of the hidden field used by gorilla/csrf.
	CSRFField = "gorilla.csrf.Token"

	// TaskField is used by submit buttons to indicate the chosen task.
	TaskField = "task"

	// TemplateRunTaskPage is the template used for task execution feedback.
	TemplateRunTaskPage Page = "domains/admin/runTaskPage.gohtml"
)
