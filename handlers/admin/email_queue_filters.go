package admin

import (
	"database/sql"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// emailFilterStatusPending filters unsent emails without errors.
	emailFilterStatusPending = "pending"
	// emailFilterStatusFailed filters unsent emails with errors.
	emailFilterStatusFailed = "failed"
	// emailFilterProviderDirect filters direct emails.
	emailFilterProviderDirect = "direct"
	// emailFilterProviderUser filters emails tied to user accounts.
	emailFilterProviderUser = "user"
	// emailFilterProviderUserless filters queued emails without users.
	emailFilterProviderUserless = "userless"
	// emailFilterAgeDay filters emails older than 24 hours.
	emailFilterAgeDay = "24h"
	// emailFilterAgeWeek filters emails older than 7 days.
	emailFilterAgeWeek = "7d"
	// emailFilterAgeMonth filters emails older than 30 days.
	emailFilterAgeMonth = "30d"
)

// EmailFilters captures filter values for the email views.
type EmailFilters struct {
	Status        string
	Provider      string
	Age           string
	CreatedBefore sql.NullTime
	LangID        int
	Role          string
}

func emailFiltersFromValues(values url.Values) EmailFilters {
	filters := EmailFilters{}
	status := normalizeEmailFilter(values.Get("status"), []string{emailFilterStatusPending, emailFilterStatusFailed})
	provider := normalizeEmailFilter(values.Get("provider"), []string{emailFilterProviderDirect, emailFilterProviderUser, emailFilterProviderUserless})
	age := normalizeEmailFilter(values.Get("age"), []string{emailFilterAgeDay, emailFilterAgeWeek, emailFilterAgeMonth})

	filters.Status = status
	filters.Provider = provider
	filters.Age = age

	if age != "" {
		if d, ok := emailFilterAgeDuration(age); ok {
			filters.CreatedBefore = sql.NullTime{Time: time.Now().Add(-d), Valid: true}
		}
	}

	langID, _ := strconv.Atoi(values.Get("lang"))
	filters.LangID = langID
	filters.Role = values.Get("role")

	return filters
}

func normalizeEmailFilter(value string, allowed []string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return value
		}
	}
	return ""
}

func emailFilterAgeDuration(age string) (time.Duration, bool) {
	switch age {
	case emailFilterAgeDay:
		return 24 * time.Hour, true
	case emailFilterAgeWeek:
		return 7 * 24 * time.Hour, true
	case emailFilterAgeMonth:
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func (f EmailFilters) StatusParam() sql.NullString {
	return sql.NullString{String: f.Status, Valid: f.Status != ""}
}

func (f EmailFilters) ProviderParam() sql.NullString {
	return sql.NullString{String: f.Provider, Valid: f.Provider != ""}
}

func (f EmailFilters) LangIDParam() sql.NullInt32 {
	return sql.NullInt32{Int32: int32(f.LangID), Valid: f.LangID != 0}
}

func (f EmailFilters) RoleParam() sql.NullString {
	return sql.NullString{String: f.Role, Valid: f.Role != ""}
}

// AuditSummary returns a summary of filters for audit logs.
func (f EmailFilters) AuditSummary() string {
	parts := make([]string, 0, 5)
	if f.Status != "" {
		parts = append(parts, "status="+f.Status)
	}
	if f.Provider != "" {
		parts = append(parts, "provider="+f.Provider)
	}
	if f.Age != "" {
		parts = append(parts, "age="+f.Age)
	}
	if f.LangID != 0 {
		parts = append(parts, "lang="+strconv.Itoa(f.LangID))
	}
	if f.Role != "" {
		parts = append(parts, "role="+f.Role)
	}
	if len(parts) == 0 {
		return "no filters"
	}
	return strings.Join(parts, ", ")
}
