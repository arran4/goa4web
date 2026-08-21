-- +goose Up
UPDATE linker SET linker_category_id = NULL WHERE linker_category_id = 0;
UPDATE linker_queue SET linker_category_id = NULL WHERE linker_category_id = 0;
UPDATE deactivated_linker SET linker_category_id = NULL WHERE linker_category_id = 0;

UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;

UPDATE faq SET faqCategories_idfaqCategories = NULL WHERE faqCategories_idfaqCategories = 0;

UPDATE imageboard SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;
UPDATE imagepost SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;
UPDATE deactivated_imageposts SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;

UPDATE schema_version SET version = 58;