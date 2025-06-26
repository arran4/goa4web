package goa4web

import db "github.com/arran4/goa4web/internal/db"

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

func New(d db.DBTX) *Queries { return db.New(d) }
