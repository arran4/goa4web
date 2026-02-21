CREATE TABLE `1_old_forumthread` (
  `idforumthread` int(10) NOT NULL AUTO_INCREMENT,
  `forumtopic_idforumtopic` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`idforumthread`),
  KEY `forumdiscussions_FKIndex1` (`forumtopic_idforumtopic`)
);

CREATE TABLE `1_old_forumtopic` (
  `idforumtopic` int(10) NOT NULL AUTO_INCREMENT,
  `forumcategory_idforumcategory` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idforumtopic`),
  KEY `forumtopic_FKIndex1` (`forumcategory_idforumcategory`)
);

CREATE TABLE `blogs` (
  `idblogs` int(10) NOT NULL AUTO_INCREMENT,
  `forumthread_idforumthread` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `blog` longtext DEFAULT NULL,
  `written` DATETIME NOT NULL DEFAULT NOW(),
  PRIMARY KEY (`idblogs`),
  KEY `blogs_FKIndex1` (`language_id`),
  KEY `blogs_FKIndex2` (`users_idusers`),
  KEY `blogs_FKIndex3` (`forumthread_idforumthread`)
);

CREATE TABLE `blogsSearch` (
  `blogs_idblogs` int(10) NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`blogs_idblogs`,`searchwordlist_idsearchwordlist`),
  KEY `blogs_has_searchwordlist_FKIndex1` (`blogs_idblogs`),
  KEY `blogs_has_searchwordlist_FKIndex2` (`searchwordlist_idsearchwordlist`)
);

CREATE TABLE `bookmarks` (
  `idbookmarks` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `list` mediumblob DEFAULT NULL,
  PRIMARY KEY (`idbookmarks`),
  KEY `bookmarks_FKIndex1` (`users_idusers`)
);

CREATE TABLE `comments` (
  `idcomments` int(10) NOT NULL AUTO_INCREMENT,
  `forumthread_idforumthread` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `written` datetime DEFAULT NULL,
  `text` longtext DEFAULT NULL,
  PRIMARY KEY (`idcomments`),
  KEY `comments_FKIndex1` (`language_id`),
  KEY `comments_FKIndex2` (`users_idusers`),
  KEY `comments_FKIndex3` (`forumthread_idforumthread`)
);

CREATE TABLE `commentsSearch` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `comments_idcomments` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`comments_idcomments`),
  KEY `searchwordlist_has_comments_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_comments_FKIndex2` (`comments_idcomments`)
);

CREATE TABLE `faq` (
  `idfaq` int(10) NOT NULL AUTO_INCREMENT,
  `faqCategories_idfaqCategories` int(10) NOT NULL DEFAULT 0,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `answer` mediumtext DEFAULT NULL,
  `question` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idfaq`),
  KEY `Table_21_FKIndex1` (`users_idusers`),
  KEY `Table_21_FKIndex2` (`language_id`),
  KEY `Table_21_FKIndex3` (`faqCategories_idfaqCategories`)
);

CREATE TABLE `faqCategories` (
  `idfaqCategories` int(10) NOT NULL AUTO_INCREMENT,
  `name` tinytext DEFAULT NULL,
  PRIMARY KEY (`idfaqCategories`)
);

CREATE TABLE `forumcategory` (
  `idforumcategory` int(10) NOT NULL AUTO_INCREMENT,
  `forumcategory_idforumcategory` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idforumcategory`),
  KEY `forumcategory_FKIndex1` (`forumcategory_idforumcategory`)
);

CREATE TABLE `forumthread` (
  `idforumthread` int(10) NOT NULL AUTO_INCREMENT,
  `firstpost` int(10) NOT NULL DEFAULT 0,
  `lastposter` int(10) NOT NULL DEFAULT 0,
  `forumtopic_idforumtopic` int(10) NOT NULL DEFAULT 0,
  `comments` int(10) DEFAULT NULL,
  `lastaddition` datetime DEFAULT NULL,
  `locked` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`idforumthread`),
  KEY `forumdiscussions_FKIndex1` (`forumtopic_idforumtopic`),
  KEY `forumthread_FKIndex2` (`lastposter`),
  KEY `forumthread_FKIndex3` (`firstpost`)
);

CREATE TABLE `forumtopic` (
  `idforumtopic` int(10) NOT NULL AUTO_INCREMENT,
  `lastposter` int(10) NOT NULL DEFAULT 0,
  `forumcategory_idforumcategory` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  `threads` int(10) DEFAULT NULL,
  `comments` int(10) DEFAULT NULL,
  `lastaddition` datetime DEFAULT NULL,
  PRIMARY KEY (`idforumtopic`),
  KEY `forumtopic_FKIndex1` (`forumcategory_idforumcategory`),
  KEY `forumtopic_FKIndex2` (`lastposter`)
);

CREATE TABLE `imageboard` (
  `idimageboard` int(10) NOT NULL AUTO_INCREMENT,
  `imageboard_idimageboard` int(10) DEFAULT NULL,
  `title` tinytext DEFAULT NULL,
  `description` tinytext DEFAULT NULL,
  `approval_required` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`idimageboard`),
  KEY `imageboard_FKIndex1` (`imageboard_idimageboard`)
);

CREATE TABLE `imagepost` (
  `idimagepost` int(10) NOT NULL AUTO_INCREMENT,
  `forumthread_idforumthread` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `imageboard_idimageboard` int(10) DEFAULT NULL,
  `posted` datetime DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  `thumbnail` tinytext DEFAULT NULL,
  `fullimage` tinytext DEFAULT NULL,
  `file_size` int(10) NOT NULL DEFAULT 0,
  `approved` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`idimagepost`),
  KEY `imagepost_FKIndex1` (`imageboard_idimageboard`),
  KEY `imagepost_FKIndex2` (`users_idusers`),
  KEY `imagepost_FKIndex3` (`forumthread_idforumthread`)
);

CREATE TABLE `imagepostSearch` (
  `imagepost_idimagepost` int(10) NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`imagepost_idimagepost`,`searchwordlist_idsearchwordlist`),
  KEY `imagepostSearch_FKIndex1` (`imagepost_idimagepost`),
  KEY `imagepostSearch_FKIndex2` (`searchwordlist_idsearchwordlist`)
);

CREATE TABLE `language` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `nameof` tinytext DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE `linker` (
  `idlinker` int(10) NOT NULL AUTO_INCREMENT,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `linkerCategory_idlinkerCategory` int(10) NOT NULL DEFAULT 0,
  `forumthread_idforumthread` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `url` tinytext DEFAULT NULL,
  `description` tinytext DEFAULT NULL,
  `listed` datetime DEFAULT NULL,
  PRIMARY KEY (`idlinker`),
  KEY `linker_FKIndex1` (`forumthread_idforumthread`),
  KEY `linker_FKIndex2` (`linkerCategory_idlinkerCategory`),
  KEY `linker_FKIndex3` (`users_idusers`),
  KEY `linker_FKIndex4` (`language_id`)
);

CREATE TABLE `linkerCategory` (
  `idlinkerCategory` int(10) NOT NULL AUTO_INCREMENT,
  `position` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `sortorder` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`idlinkerCategory`)
);

CREATE TABLE `linkerQueue` (
  `idlinkerQueue` int(10) NOT NULL AUTO_INCREMENT,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `linkerCategory_idlinkerCategory` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `url` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idlinkerQueue`),
  KEY `linkerQueue_FKIndex1` (`linkerCategory_idlinkerCategory`),
  KEY `linkerQueue_FKIndex2` (`users_idusers`),
  KEY `linkerQueue_FKIndex3` (`language_id`)
);

CREATE TABLE `linkerSearch` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `linker_idlinker` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`linker_idlinker`),
  KEY `searchwordlist_has_linker_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_linker_FKIndex2` (`linker_idlinker`)
);

CREATE TABLE `permissions` (
  `idpermissions` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `section` tinytext DEFAULT NULL,
  `level` tinyblob DEFAULT NULL,
  PRIMARY KEY (`idpermissions`),
  KEY `permissions_FKIndex1` (`users_idusers`)
);

CREATE TABLE `preferences` (
  `idpreferences` int(10) NOT NULL AUTO_INCREMENT,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `emailforumupdates` tinyint(1) DEFAULT 0,
  `page_size` int(10) NOT NULL DEFAULT 15,
  PRIMARY KEY (`idpreferences`),
  KEY `preferences_FKIndex1` (`users_idusers`),
  KEY `preferences_FKIndex2` (`language_id`)
);

CREATE TABLE `searchwordlist` (
  `idsearchwordlist` int(10) NOT NULL AUTO_INCREMENT,
  `word` tinytext DEFAULT NULL,
  PRIMARY KEY (`idsearchwordlist`)
);

CREATE TABLE `searchwordlist_has_linker` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `linker_idlinker` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`linker_idlinker`),
  KEY `searchwordlist_has_linker_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_linker_FKIndex2` (`linker_idlinker`)
);

CREATE TABLE `siteNews` (
  `idsiteNews` int(10) NOT NULL AUTO_INCREMENT,
  `forumthread_idforumthread` int(10) NOT NULL DEFAULT 0,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `news` longtext DEFAULT NULL,
  `occured` datetime DEFAULT NULL,
  PRIMARY KEY (`idsiteNews`),
  KEY `siteNews_FKIndex1` (`users_idusers`),
  KEY `siteNews_FKIndex2` (`language_id`),
  KEY `siteNews_FKIndex3` (`forumthread_idforumthread`)
);

CREATE TABLE `siteNewsSearch` (
  `siteNews_idsiteNews` int(10) NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`siteNews_idsiteNews`,`searchwordlist_idsearchwordlist`),
  KEY `siteNews_has_searchwordlist_FKIndex1` (`siteNews_idsiteNews`),
  KEY `siteNews_has_searchwordlist_FKIndex2` (`searchwordlist_idsearchwordlist`)
);

CREATE TABLE `topicrestrictions` (
  `forumtopic_idforumtopic` int(10) NOT NULL DEFAULT 0,
  `viewlevel` int(10) DEFAULT NULL,
  `replylevel` int(10) DEFAULT NULL,
  `newthreadlevel` int(10) DEFAULT NULL,
  `seelevel` int(10) DEFAULT NULL,
  `invitelevel` int(10) DEFAULT NULL,
  `readlevel` int(10) DEFAULT NULL,
  `modlevel` int(10) DEFAULT NULL,
  `adminlevel` int(10) DEFAULT NULL,
  PRIMARY KEY (`forumtopic_idforumtopic`),
  KEY `threadrestrictions_FKIndex1` (`forumtopic_idforumtopic`)
);

CREATE TABLE `userlang` (
  `iduserlang` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `language_id` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`iduserlang`),
  KEY `userpref_FKIndex1` (`language_id`),
  KEY `userpref_FKIndex2` (`users_idusers`)
);

CREATE TABLE `users` (
  `idusers` int(10) NOT NULL AUTO_INCREMENT,
  `email` tinytext DEFAULT NULL,
  `passwd` tinytext DEFAULT NULL,
  `passwd_algorithm` tinytext DEFAULT NULL,
  `username` tinytext DEFAULT NULL,
  PRIMARY KEY (`idusers`)
);

CREATE TABLE `userstopiclevel` (
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `forumtopic_idforumtopic` int(10) NOT NULL DEFAULT 0,
  `level` int(10) DEFAULT NULL,
  `invitemax` int(10) DEFAULT NULL,
  `expires_at` datetime DEFAULT NULL,
  PRIMARY KEY (`users_idusers`,`forumtopic_idforumtopic`),
  KEY `users_has_forumtopic_FKIndex1` (`users_idusers`),
  KEY `users_has_forumtopic_FKIndex2` (`forumtopic_idforumtopic`)
);

CREATE TABLE `writing` (
  `idwriting` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `forumthread_idforumthread` int(10) NOT NULL DEFAULT 0,
  `language_id` int(10) NOT NULL DEFAULT 0,
  `writingCategory_idwritingCategory` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `published` datetime DEFAULT NULL,
  `writting` longtext DEFAULT NULL,
  `abstract` mediumtext DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`idwriting`),
  KEY `writing_FKIndex1` (`writingCategory_idwritingCategory`),
  KEY `writing_FKIndex2` (`language_id`),
  KEY `writing_FKIndex3` (`forumthread_idforumthread`),
  KEY `writing_FKIndex4` (`users_idusers`)
);

CREATE TABLE `writingCategory` (
  `idwritingCategory` int(10) NOT NULL AUTO_INCREMENT,
  `writingCategory_idwritingCategory` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `description` tinytext DEFAULT NULL,
  PRIMARY KEY (`idwritingCategory`),
  KEY `writingCategory_FKIndex1` (`writingCategory_idwritingCategory`)
);

CREATE TABLE `writingSearch` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `writing_idwriting` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`writing_idwriting`),
  KEY `searchwordlist_has_writing_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_writing_FKIndex2` (`writing_idwriting`)
);

CREATE TABLE `writtingApprovedUsers` (
  `writing_idwriting` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `readdoc` tinyint(1) DEFAULT NULL,
  `editdoc` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`writing_idwriting`,`users_idusers`),
  KEY `writing_has_users_FKIndex1` (`writing_idwriting`),
  KEY `writing_has_users_FKIndex2` (`users_idusers`)
);

-- Track schema upgrades.
CREATE TABLE IF NOT EXISTS `schema_version` (
  `version` int NOT NULL
);

-- Store subscribed users for notifications.
CREATE TABLE IF NOT EXISTS `subscriptions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `item_type` varchar(32) NOT NULL,
  `target_id` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
);

-- Queue outbound emails.
CREATE TABLE IF NOT EXISTS `pending_emails` (
  `id` int NOT NULL AUTO_INCREMENT,
  `to_email` text NOT NULL,
  `subject` text NOT NULL,
  `body` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sent_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

-- Internal notification list.
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `link` text,
  `message` text,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `read_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

-- Track active user sessions.
CREATE TABLE IF NOT EXISTS `sessions` (
  `session_id` varchar(128) NOT NULL,
  `users_idusers` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`session_id`),
  KEY `sessions_user_idx` (`users_idusers`)
);

-- Site announcements referencing promoted news posts.
CREATE TABLE IF NOT EXISTS `site_announcements` (
  `id` int NOT NULL AUTO_INCREMENT,
  `site_news_id` int NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `site_announcements_news_idx` (`site_news_id`)
);

-- Track failed login attempts.
CREATE TABLE IF NOT EXISTS `login_attempts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `ip_address` varchar(45) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
);

-- IP addresses banned from accessing the site.
CREATE TABLE IF NOT EXISTS `banned_ips` (
  `id` int NOT NULL AUTO_INCREMENT,
  `ip_net` varchar(50) NOT NULL,
  `reason` text,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expires_at` datetime DEFAULT NULL,
  `canceled_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `banned_ips_ip_idx` (`ip_net`)
);

-- Optional template overrides for dynamic content.
CREATE TABLE IF NOT EXISTS `template_overrides` (
  `name` varchar(128) NOT NULL,
  `body` text NOT NULL,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`name`)
);

-- Audit log of administrative actions.
CREATE TABLE IF NOT EXISTS `audit_log` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `action` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `audit_log_user_idx` (`users_idusers`)
);

