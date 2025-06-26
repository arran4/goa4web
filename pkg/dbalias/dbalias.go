package dbalias

import db "github.com/arran4/goa4web/internal/db"

// These aliases expose selected internal db types.
type (
	DBTX                                                = db.DBTX
	Queries                                             = db.Queries
	GetActiveAnnouncementWithNewsRow                    = db.GetActiveAnnouncementWithNewsRow
	GetPermissionsByUserIdAndSectionAndSectionAllParams = db.GetPermissionsByUserIdAndSectionAndSectionAllParams
	InsertAuditLogParams                                = db.InsertAuditLogParams
	Language                                            = db.Language
	Linker                                              = db.Linker
	User                                                = db.User
)

// New returns a new Queries instance using the given database.
func New(d db.DBTX) *Queries { return db.New(d) }
