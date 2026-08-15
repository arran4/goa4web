-- +goose Up
UPDATE grants SET section = 'writing' WHERE section = 'writings';