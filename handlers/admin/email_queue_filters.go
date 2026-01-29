package admin

import (
	"database/sql"
	"net/url"
	"strings"
	"time"
)

const (
	// emailQueueStatusPending filters unsent emails without errors.
	emailQueueStatusPending = "pending"
	// emailQueueStatusFailed filters unsent emails with errors.
	emailQueueStatusFailed = "failed"
	// emailQueueProviderDirect filters direct emails.
	emailQueueProviderDirect = "direct"
	// emailQueueProviderUser filters emails tied to user accounts.
	emailQueueProviderUser = "user"
	// emailQueueProviderUserless filters queued emails without users.
	emailQueueProviderUserless = "userless"
	// emailQueueAgeDay filters emails older than 24 hours.
	emailQueueAgeDay = "24h"
	// emailQueueAgeWeek filters emails older than 7 days.
	emailQueueAgeWeek = "7d"
	// emailQueueAgeMonth filters emails older than 30 days.
	emailQueueAgeMonth = "30d"
)

// EmailQueueFilters captures filter values for the queue view.
type EmailQueueFilters struct {
	Status        string
	Provider      string
	Age           string
	CreatedBefore sql.NullTime
}

func emailQueueFiltersFromValues(values url.Values) EmailQueueFilters {
	filters := EmailQueueFilters{}
	status := normalizeEmailQueueFilter(values.Get("status"), []string{emailQueueStatusPending, emailQueueStatusFailed})
	provider := normalizeEmailQueueFilter(values.Get("provider"), []string{emailQueueProviderDirect, emailQueueProviderUser, emailQueueProviderUserless})
	age := normalizeEmailQueueFilter(values.Get("age"), []string{emailQueueAgeDay, emailQueueAgeWeek, emailQueueAgeMonth})
	filters.Status = status
	filters.Provider = provider
	filters.Age = age
	if age != "" {
		if d, ok := emailQueueAgeDuration(age); ok {
			filters.CreatedBefore = sql.NullTime{Time: time.Now().Add(-d), Valid: true}
		}
	}
	return filters
}

func normalizeEmailQueueFilter(value string, allowed []string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return value
		}
	}
	return ""
}

func emailQueueAgeDuration(age string) (time.Duration, bool) {
	switch age {
	case emailQueueAgeDay:
		return 24 * time.Hour, true
	case emailQueueAgeWeek:
		return 7 * 24 * time.Hour, true
	case emailQueueAgeMonth:
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func (f EmailQueueFilters) StatusParam() sql.NullString {
	return sql.NullString{String: f.Status, Valid: f.Status != ""}
}

func (f EmailQueueFilters) ProviderParam() sql.NullString {
	return sql.NullString{String: f.Provider, Valid: f.Provider != ""}
}

// AuditSummary returns a summary of filters for audit logs.
func (f EmailQueueFilters) AuditSummary() string {
	parts := make([]string, 0, 3)
	if f.Status != "" {
		parts = append(parts, "status="+f.Status)
	}
	if f.Provider != "" {
		parts = append(parts, "provider="+f.Provider)
	}
	if f.Age != "" {
		parts = append(parts, "age="+f.Age)
	}
	if len(parts) == 0 {
		return "no filters"
	}
	return strings.Join(parts, ", ")
}
