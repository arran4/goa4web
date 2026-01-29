package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

type requestActionResult struct {
	ID          int32  `json:"id"`
	Status      string `json:"status,omitempty"`
	AutoComment string `json:"auto_comment,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func updateRequestStatus(ctx context.Context, queries *db.Queries, requestID int32, status string, comment string) (*requestActionResult, error) {
	req, err := queries.AdminGetRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}
	if err := queries.AdminUpdateRequestStatus(ctx, db.AdminUpdateRequestStatusParams{Status: status, ID: requestID}); err != nil {
		return nil, fmt.Errorf("update request status: %w", err)
	}
	auto := fmt.Sprintf("status changed to %s", status)
	if err := queries.AdminInsertRequestComment(ctx, db.AdminInsertRequestCommentParams{RequestID: requestID, Comment: auto}); err != nil {
		return nil, fmt.Errorf("insert request comment: %w", err)
	}
	if comment != "" {
		if err := queries.AdminInsertRequestComment(ctx, db.AdminInsertRequestCommentParams{RequestID: requestID, Comment: comment}); err != nil {
			return nil, fmt.Errorf("insert request comment: %w", err)
		}
	}
	if err := queries.InsertAdminUserComment(ctx, db.InsertAdminUserCommentParams{UsersIdusers: req.UsersIdusers, Comment: auto}); err != nil {
		return nil, fmt.Errorf("insert admin user comment: %w", err)
	}
	return &requestActionResult{
		ID:          requestID,
		Status:      status,
		AutoComment: auto,
		Comment:     comment,
	}, nil
}

func emitRequestActionJSON(result *requestActionResult) error {
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
	return nil
}
