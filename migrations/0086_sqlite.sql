-- +goose Up
ALTER TABLE preferences ADD COLUMN image_safe_dimension TEXT;