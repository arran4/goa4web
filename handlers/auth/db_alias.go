package auth

import db "github.com/arran4/goa4web/internal/db"

type (
	DBTX                     = db.DBTX
	Queries                  = db.Queries
	InsertLoginAttemptParams = db.InsertLoginAttemptParams
	User                     = db.User
)

func New(d db.DBTX) *Queries { return db.New(d) }
