package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserDeactivateCmd_Run(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	root := &rootCmd{db: db}
	uc := &userCmd{rootCmd: root}
	cmd, err := parseUserDeactivateCmd(uc, []string{"--id", "1"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	mock.ExpectExec("INSERT INTO deactivated_users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO deactivated_comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM permissions").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM sessions").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestUserRestoreCmd_Run(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	root := &rootCmd{db: db}
	uc := &userCmd{rootCmd: root}
	cmd, err := parseUserRestoreCmd(uc, []string{"--id", "1"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	mock.ExpectExec("UPDATE users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM deactivated_users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM deactivated_comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
