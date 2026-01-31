package admin

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/handlers"
)

func emailFiltersFromRequest(r *http.Request) EmailFilters {
	return emailFiltersFromValues(r.URL.Query())
}

func buildEmailStatusMap(r *http.Request, pageIDs []int32) map[int32]string {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		return nil
	}
	label := emailStatusLabel(status)
	if label == "" {
		label = status
	}
	scope := r.URL.Query().Get("scope")
	if scope == "filtered" {
		statuses := make(map[int32]string, len(pageIDs))
		for _, id := range pageIDs {
			statuses[id] = label
		}
		return statuses
	}
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		return nil
	}
	idSet := make(map[int32]struct{})
	for _, part := range strings.Split(idsParam, ",") {
		if part == "" {
			continue
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		idSet[int32(value)] = struct{}{}
	}
	if len(idSet) == 0 {
		return nil
	}
	statuses := make(map[int32]string)
	for _, id := range pageIDs {
		if _, ok := idSet[id]; ok {
			statuses[id] = label
		}
	}
	return statuses
}

func emailStatusLabel(status string) string {
	switch status {
	case "resent":
		return "Resent"
	case "retry":
		return "Queued for retry"
	default:
		return status
	}
}

func buildEmailTaskRedirect(r *http.Request, status string, scope string, ids []int32) handlers.RefreshDirectHandler {
	vals := url.Values{}
	for key, values := range r.URL.Query() {
		vals[key] = append([]string(nil), values...)
	}
	if status != "" {
		vals.Set("status", status)
	} else {
		vals.Del("status")
	}
	if scope != "" {
		vals.Set("scope", scope)
	} else {
		vals.Del("scope")
	}
	if scope == "filtered" {
		vals.Del("ids")
	} else if len(ids) > 0 {
		vals.Set("ids", joinEmailIDs(ids))
	} else {
		vals.Del("ids")
	}
	target := r.URL.Path
	if encoded := vals.Encode(); encoded != "" {
		target += "?" + encoded
	}
	return handlers.RefreshDirectHandler{TargetURL: target}
}

func joinEmailIDs(ids []int32) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.Itoa(int(id)))
	}
	return strings.Join(parts, ",")
}

func emailFilterSummary(prefix string, f EmailFilters) string {
	s := f.AuditSummary()
	if prefix != "" {
		if s == "no filters" {
			return prefix
		}
		return prefix + ", " + s
	}
	return s
}
