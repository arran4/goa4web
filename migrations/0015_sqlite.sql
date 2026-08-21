-- +goose Up
ALTER TABLE worker_errors RENAME TO dead_letters;