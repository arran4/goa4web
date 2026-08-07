-- +goose Up
-- Normalize writing section name
UPDATE grants SET section = 'writing' WHERE section = 'writings';

-- Update schema version
