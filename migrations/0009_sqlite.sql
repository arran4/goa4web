-- +goose Up
ALTER TABLE writing
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

ALTER TABLE blogs
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

ALTER TABLE imagepost
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

ALTER TABLE linker
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

CREATE TABLE IF NOT EXISTS deactivated_writings (
idwriting INT NOT NULL,
users_idusers INT NOT NULL,
forumthread_idforumthread INT NOT NULL,
language_idlanguage INT NOT NULL,
writingCategory_idwritingCategory INT NOT NULL,
title TEXT,
published DATETIME,
writting TEXT,
abstract TEXT,
private INTEGER DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idwriting)
);

CREATE TABLE IF NOT EXISTS deactivated_blogs (
idblogs INT NOT NULL,
forumthread_idforumthread INT NOT NULL,
users_idusers INT NOT NULL,
language_idlanguage INT NOT NULL,
blog TEXT,
written DATETIME,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idblogs)
);

CREATE TABLE IF NOT EXISTS deactivated_imageposts (
idimagepost INT NOT NULL,
forumthread_idforumthread INT NOT NULL,
users_idusers INT NOT NULL,
imageboard_idimageboard INT NOT NULL,
posted DATETIME,
description TEXT,
thumbnail TEXT,
fullimage TEXT,
file_size INT NOT NULL,
approved INTEGER DEFAULT 0,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idimagepost)
);

CREATE TABLE IF NOT EXISTS deactivated_linker (
idlinker INT NOT NULL,
language_idlanguage INT NOT NULL,
users_idusers INT NOT NULL,
linkerCategory_idlinkerCategory INT NOT NULL,
forumthread_idforumthread INT NOT NULL,
title TEXT,
url TEXT,
description TEXT,
listed DATETIME,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idlinker)
);