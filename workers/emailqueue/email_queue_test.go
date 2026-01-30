package emailqueue

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

func TestProcessPendingEmail_NilProvider_IncrementsErrorCount(t *testing.T) {
	q := &db.QuerierStub{}
	cfg := &config.RuntimeConfig{
		EmailEnabled: true,
	}

	emailAddr := "verified@example.com"
	e := &db.SystemListPendingEmailsRow{
		ID:          1,
		ToUserID:    sql.NullInt32{Valid: false},
		Body:        fmt.Sprintf("To: %s\r\nSubject: Test\r\n\r\nBody", emailAddr),
		DirectEmail: true,
	}

	listCalled := false
	q.SystemListPendingEmailsFn = func(ctx context.Context, arg db.SystemListPendingEmailsParams) ([]*db.SystemListPendingEmailsRow, error) {
		listCalled = true
		return []*db.SystemListPendingEmailsRow{e}, nil
	}

	q.GetUserEmailByEmailFn = func(ctx context.Context, email string) (*db.UserEmail, error) {
		return &db.UserEmail{
			VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}, nil
	}

	incrementCalled := false
	q.SystemIncrementPendingEmailErrorFn = func(ctx context.Context, id int32) error {
		if id != e.ID {
			t.Errorf("Expected increment for ID %d, got %d", e.ID, id)
		}
		incrementCalled = true
		return nil
	}

	markSentCalled := false
	q.SystemMarkPendingEmailSentFn = func(ctx context.Context, id int32) error {
		markSentCalled = true
		return nil
	}

	// Call with nil provider
	result := ProcessPendingEmail(context.Background(), q, nil, nil, cfg)

	if !result {
		t.Error("Expected ProcessPendingEmail to return true (work done/waiting) when provider is nil")
	}

	if !listCalled {
		t.Error("Expected SystemListPendingEmails to be called")
	}

	if !incrementCalled {
		t.Error("Expected SystemIncrementPendingEmailError to be called")
	}

	if markSentCalled {
		t.Error("Expected SystemMarkPendingEmailSent NOT to be called")
	}
}

func TestResolveQueuedEmailAddress_DirectEmail_VerifiedUser_Success(t *testing.T) {
	q := &db.QuerierStub{}
	cfg := &config.RuntimeConfig{}

	emailAddr := "verified@example.com"

	// Mock GetUserEmailByEmail to return a verified user
	q.GetUserEmailByEmailFn = func(ctx context.Context, email string) (*db.UserEmail, error) {
		if email == emailAddr {
			return &db.UserEmail{
				VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
			}, nil
		}
		return nil, fmt.Errorf("not found")
	}

	// Mock SystemListPendingEmailsRow
	e := &db.SystemListPendingEmailsRow{
		ID:          1,
		ToUserID:    sql.NullInt32{Valid: false},
		Body:        fmt.Sprintf("To: %s\r\nSubject: Test\r\n\r\nBody", emailAddr),
		DirectEmail: true,
	}

	addr, err := ResolveQueuedEmailAddress(context.Background(), q, cfg, e)
	if err != nil {
		t.Fatalf("Expected success for verified user with DirectEmail, got error: %v", err)
	}
	if addr.Address != emailAddr {
		t.Errorf("Expected address %s, got %s", emailAddr, addr.Address)
	}
}

func TestResolveQueuedEmailAddress_DirectEmail_NonUser_Fails(t *testing.T) {
	q := &db.QuerierStub{}
	cfg := &config.RuntimeConfig{}

	emailAddr := "nonuser@example.com"

	// Mock GetUserEmailByEmail to return not found
	q.GetUserEmailByEmailFn = func(ctx context.Context, email string) (*db.UserEmail, error) {
		return nil, fmt.Errorf("not found")
	}

	// Mock SystemListPendingEmailsRow
	e := &db.SystemListPendingEmailsRow{
		ID:          1,
		ToUserID:    sql.NullInt32{Valid: false},
		Body:        fmt.Sprintf("To: %s\r\nSubject: Test\r\n\r\nBody", emailAddr),
		DirectEmail: true,
	}

	_, err := ResolveQueuedEmailAddress(context.Background(), q, cfg, e)
	if err == nil {
		t.Fatal("Expected error for non-user with DirectEmail, got nil")
	}
	expectedErr := fmt.Sprintf("no verification record for %s", emailAddr)
	if err.Error() != expectedErr {
		t.Fatalf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestResolveQueuedEmailAddress_DirectEmail_UnverifiedUser_Success(t *testing.T) {
	q := &db.QuerierStub{}
	cfg := &config.RuntimeConfig{}

	emailAddr := "unverified@example.com"

	// Mock GetUserEmailByEmail to return an unverified user with valid verification code
	q.GetUserEmailByEmailFn = func(ctx context.Context, email string) (*db.UserEmail, error) {
		if email == emailAddr {
			return &db.UserEmail{
				VerifiedAt:            sql.NullTime{Valid: false},
				LastVerificationCode:  sql.NullString{String: "code", Valid: true},
				VerificationExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
			}, nil
		}
		return nil, fmt.Errorf("not found")
	}

	// Mock SystemListPendingEmailsRow
	e := &db.SystemListPendingEmailsRow{
		ID:          1,
		ToUserID:    sql.NullInt32{Valid: false},
		Body:        fmt.Sprintf("To: %s\r\nSubject: Test\r\n\r\nBody", emailAddr),
		DirectEmail: true,
	}

	addr, err := ResolveQueuedEmailAddress(context.Background(), q, cfg, e)
	if err != nil {
		t.Fatalf("Expected success for unverified user with DirectEmail, got error: %v", err)
	}
	if addr.Address != emailAddr {
		t.Errorf("Expected address %s, got %s", emailAddr, addr.Address)
	}
}
