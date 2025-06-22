-- Add search table for image posts
CREATE TABLE IF NOT EXISTS `imagepostSearch` (
  `imagepost_idimagepost` INT NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`imagepost_idimagepost`, `searchwordlist_idsearchwordlist`),
  KEY `imagepostSearch_FKIndex1` (`imagepost_idimagepost`),
  KEY `imagepostSearch_FKIndex2` (`searchwordlist_idsearchwordlist`)
);
