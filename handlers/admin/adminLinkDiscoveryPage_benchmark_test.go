package admin

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func BenchmarkRegexCompileInLoop(b *testing.B) {
	rows := make([]*db.GetAllSiteNewsForIndexRow, 100)
	for i := 0; i < 100; i++ {
		rows[i] = &db.GetAllSiteNewsForIndexRow{
			News: sql.NullString{String: "Some text with a link http://example.com/foo and more text.", Valid: true},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := regexp.MustCompile(`https?://[^\s"']+`)
		for _, row := range rows {
			if row.News.Valid {
				_ = re.FindAllString(row.News.String, -1)
			}
		}
	}
}

func BenchmarkRegexCompileOutside(b *testing.B) {
	rows := make([]*db.GetAllSiteNewsForIndexRow, 100)
	for i := 0; i < 100; i++ {
		rows[i] = &db.GetAllSiteNewsForIndexRow{
			News: sql.NullString{String: "Some text with a link http://example.com/foo and more text.", Valid: true},
		}
	}

	re := regexp.MustCompile(`https?://[^\s"']+`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, row := range rows {
			if row.News.Valid {
				_ = re.FindAllString(row.News.String, -1)
			}
		}
	}
}
