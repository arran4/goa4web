CREATE TABLE IF NOT EXISTS `imagepostSearch` (
  `imagepost_idimagepost` INT NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`imagepost_idimagepost`, `searchwordlist_idsearchwordlist`),
  KEY `imagepostSearch_FKIndex1` (`imagepost_idimagepost`),
  KEY `imagepostSearch_FKIndex2` (`searchwordlist_idsearchwordlist`)
);

-- Record upgrade to schema version 4
UPDATE schema_version SET version = 4 WHERE version = 3;
