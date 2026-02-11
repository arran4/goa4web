package db

import (
	"reflect"
	"testing"
)

func TestEmailsByUserID(t *testing.T) {
	tests := []struct {
		name     string
		rows     []*GetVerifiedUserEmailsRow
		expected map[int32][]string
	}{
		{
			name:     "Empty input",
			rows:     []*GetVerifiedUserEmailsRow{},
			expected: map[int32][]string{},
		},
		{
			name: "Single user, single email",
			rows: []*GetVerifiedUserEmailsRow{
				{UserID: 1, Email: "user1@example.com"},
			},
			expected: map[int32][]string{
				1: {"user1@example.com"},
			},
		},
		{
			name: "Single user, multiple emails",
			rows: []*GetVerifiedUserEmailsRow{
				{UserID: 1, Email: "user1@example.com"},
				{UserID: 1, Email: "user1.backup@example.com"},
			},
			expected: map[int32][]string{
				1: {"user1@example.com", "user1.backup@example.com"},
			},
		},
		{
			name: "Multiple users, multiple emails",
			rows: []*GetVerifiedUserEmailsRow{
				{UserID: 1, Email: "user1@example.com"},
				{UserID: 2, Email: "user2@example.com"},
				{UserID: 1, Email: "user1.backup@example.com"},
			},
			expected: map[int32][]string{
				1: {"user1@example.com", "user1.backup@example.com"},
				2: {"user2@example.com"},
			},
		},
		{
			name: "Nil element in rows",
			rows: []*GetVerifiedUserEmailsRow{
				{UserID: 1, Email: "user1@example.com"},
				nil,
			},
			expected: map[int32][]string{
				1: {"user1@example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EmailsByUserID(tt.rows)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("EmailsByUserID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPrimaryEmail(t *testing.T) {
	tests := []struct {
		name     string
		emails   []string
		expected string
	}{
		{
			name:     "Empty slice",
			emails:   []string{},
			expected: "",
		},
		{
			name:     "Single email",
			emails:   []string{"test@example.com"},
			expected: "test@example.com",
		},
		{
			name:     "Multiple emails",
			emails:   []string{"primary@example.com", "secondary@example.com"},
			expected: "primary@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrimaryEmail(tt.emails)
			if result != tt.expected {
				t.Errorf("PrimaryEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}
