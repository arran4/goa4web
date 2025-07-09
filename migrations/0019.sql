CREATE TABLE IF NOT EXISTS uploaded_images (
    iduploadedimage INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    path TINYTEXT,
    thumbnail TINYTEXT,
    file_size INT NOT NULL,
    uploaded DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (iduploadedimage),
    KEY uploaded_images_user_idx (users_idusers)
);

UPDATE schema_version SET version = 19 WHERE version = 18;
