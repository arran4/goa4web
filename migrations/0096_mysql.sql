-- +goose Up
ALTER TABLE external_links MODIFY url varchar(2048) NOT NULL;

-- +goose Down
ALTER TABLE external_links MODIFY url tinytext NOT NULL;
