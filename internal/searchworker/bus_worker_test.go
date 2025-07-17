package searchworker

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestBusWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := eventbus.NewBus()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)

	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").
		WithArgs("hello").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT IGNORE INTO searchwordlist").
		WithArgs("world").
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT IGNORE INTO comments_search").
		WithArgs(int32(5), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT IGNORE INTO comments_search").
		WithArgs(int32(5), int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	go BusWorker(ctx, bus, q)

	bus.Publish(eventbus.Event{Data: map[string]any{
		"search_text":  "Hello world Hello",
		"search_table": "forum",
		"search_id":    int64(5),
	}})

	time.Sleep(10 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
