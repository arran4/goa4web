CREATE TABLE IF NOT EXISTS external_links (
    id INT NOT NULL AUTO_INCREMENT,
    url tinytext NOT NULL,
    clicks INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by INT DEFAULT NULL,
    card_title tinytext,
    card_description tinytext,
    card_image tinytext,
    card_image_cache tinytext,
    favicon_cache tinytext,
    PRIMARY KEY(id),
    UNIQUE KEY external_links_url_idx (url(255))
);

UPDATE schema_version SET version = 51 WHERE version = 50;
