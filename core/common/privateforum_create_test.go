package common

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestCreatePrivateTopic(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	queries := db.New(mockDB)
	cd := &CoreData{
		queries: queries,
		ctx:     context.Background(),
		UserID:  1,
	}

	params := CreatePrivateTopicParams{
		CreatorID:      1,
		ParticipantIDs: []int32{1, 2},
		Title:          "Test Topic",
		Description:    "Test Description",
	}

	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO forumtopic").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO forumthread").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO comments").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectQuery("INSERT INTO grants").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	topicID, err := cd.CreatePrivateTopic(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if topicID != 1 {
		t.Errorf("expected topic ID 1, got %d", topicID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
