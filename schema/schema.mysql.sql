CREATE TABLE `blogs` (
  `idblogs` int(10) NOT NULL AUTO_INCREMENT,
  `forumthread_id` int(10) DEFAULT NULL,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `language_idlanguage` int(10) DEFAULT NULL,
  `blog` longtext DEFAULT NULL,
  `written` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `last_index` datetime DEFAULT NULL,
  PRIMARY KEY (`idblogs`),
  KEY `blogs_FKIndex1` (`language_idlanguage`),
  KEY `blogs_FKIndex2` (`users_idusers`),
  KEY `blogs_FKIndex3` (`forumthread_id`)
);

CREATE TABLE `blogs_search` (
  `blog_id` int(10) NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `word_count` int(10) NOT NULL DEFAULT 1,
  PRIMARY KEY (`blog_id`,`searchwordlist_idsearchwordlist`),
  KEY `blogs_has_searchwordlist_FKIndex1` (`blog_id`),
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
  `forumthread_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `language_idlanguage` int(10) DEFAULT NULL,
  `written` datetime DEFAULT NULL,
  `text` longtext DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `last_index` datetime DEFAULT NULL,
  PRIMARY KEY (`idcomments`),
  KEY `comments_FKIndex1` (`language_idlanguage`),
  KEY `comments_FKIndex2` (`users_idusers`),
  KEY `comments_FKIndex3` (`forumthread_id`)
);

CREATE TABLE `comments_search` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `comment_id` int(10) NOT NULL DEFAULT 0,
  `word_count` int(10) NOT NULL DEFAULT 1,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`comment_id`),
  KEY `searchwordlist_has_comments_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_comments_FKIndex2` (`comment_id`)
);

CREATE TABLE `faq` (
  `idfaq` int(10) NOT NULL AUTO_INCREMENT,
  `faqCategories_idfaqCategories` int(10) DEFAULT NULL,
  `language_idlanguage` int(10) DEFAULT NULL,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `answer` mediumtext DEFAULT NULL,
  `question` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idfaq`),
  KEY `Table_21_FKIndex1` (`users_idusers`),
  KEY `Table_21_FKIndex2` (`language_idlanguage`),
  KEY `Table_21_FKIndex3` (`faqCategories_idfaqCategories`)
);

CREATE TABLE `faq_categories` (
  `idfaqCategories` int(10) NOT NULL AUTO_INCREMENT,
  `name` tinytext DEFAULT NULL,
  PRIMARY KEY (`idfaqCategories`)
);

CREATE TABLE IF NOT EXISTS `faq_revisions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `faq_id` int NOT NULL,
  `users_idusers` int NOT NULL,
  `question` mediumtext,
  `answer` mediumtext,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `faq_revisions_faq_idx` (`faq_id`)
);

CREATE TABLE `forumcategory` (
  `idforumcategory` int(10) NOT NULL AUTO_INCREMENT,
  `forumcategory_idforumcategory` int(10) NOT NULL DEFAULT 0,
  `language_idlanguage` int(10) DEFAULT NULL,
  `title` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idforumcategory`),
  KEY `forumcategory_FKIndex1` (`forumcategory_idforumcategory`),
  KEY `forumcategory_FKIndex2` (`language_idlanguage`)
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
  `language_idlanguage` int(10) DEFAULT NULL,
  `title` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  `threads` int(10) DEFAULT NULL,
  `comments` int(10) DEFAULT NULL,
  `lastaddition` datetime DEFAULT NULL,
  `handler` varchar(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`idforumtopic`),
  KEY `forumtopic_FKIndex1` (`forumcategory_idforumcategory`),
  KEY `forumtopic_FKIndex2` (`lastposter`),
  KEY `forumtopic_FKIndex3` (`language_idlanguage`)
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
  `forumthread_id` int(10) NOT NULL DEFAULT 0,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `imageboard_idimageboard` int(10) DEFAULT NULL,
  `posted` datetime DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  `thumbnail` tinytext DEFAULT NULL,
  `fullimage` tinytext DEFAULT NULL,
  `file_size` int(10) NOT NULL DEFAULT 0,
  `approved` tinyint(1) NOT NULL DEFAULT 0,
  `deleted_at` datetime DEFAULT NULL,
  `last_index` datetime DEFAULT NULL,
  PRIMARY KEY (`idimagepost`),
  KEY `imagepost_FKIndex1` (`imageboard_idimageboard`),
  KEY `imagepost_FKIndex2` (`users_idusers`),
  KEY `imagepost_FKIndex3` (`forumthread_id`)
);


CREATE TABLE `imagepost_search` (
  `image_post_id` int(10) NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `word_count` int(10) NOT NULL DEFAULT 1,
  PRIMARY KEY (`image_post_id`,`searchwordlist_idsearchwordlist`),
  KEY `imagepostSearch_FKIndex1` (`image_post_id`),
  KEY `imagepostSearch_FKIndex2` (`searchwordlist_idsearchwordlist`)
);

CREATE TABLE `language` (
  `idlanguage` int(10) NOT NULL AUTO_INCREMENT,
  `nameof` tinytext DEFAULT NULL,
  PRIMARY KEY (`idlanguage`)
);

CREATE TABLE `linker` (
  `idlinker` int(10) NOT NULL AUTO_INCREMENT,
  `language_idlanguage` int(10) DEFAULT NULL,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `linker_category_id` int(10) DEFAULT NULL,
  `forumthread_id` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `url` tinytext DEFAULT NULL,
  `description` tinytext DEFAULT NULL,
  `listed` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `last_index` datetime DEFAULT NULL,
  PRIMARY KEY (`idlinker`),
  KEY `linker_FKIndex1` (`forumthread_id`),
  KEY `linker_FKIndex2` (`linker_category_id`),
  KEY `linker_FKIndex3` (`users_idusers`),
  KEY `linker_FKIndex4` (`language_idlanguage`)
);

CREATE TABLE `linker_category` (
  `idlinkerCategory` int(10) NOT NULL AUTO_INCREMENT,
  `position` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `sortorder` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`idlinkerCategory`)
);

CREATE TABLE `linker_queue` (
  `idlinkerQueue` int(10) NOT NULL AUTO_INCREMENT,
  `language_idlanguage` int(10) DEFAULT NULL,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `linker_category_id` int(10) DEFAULT NULL,
  `title` tinytext DEFAULT NULL,
  `url` tinytext DEFAULT NULL,
  `description` mediumtext DEFAULT NULL,
  PRIMARY KEY (`idlinkerQueue`),
  KEY `linkerQueue_FKIndex1` (`linker_category_id`),
  KEY `linkerQueue_FKIndex2` (`users_idusers`),
  KEY `linkerQueue_FKIndex3` (`language_idlanguage`)
);

CREATE TABLE `linker_search` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `linker_id` int(10) NOT NULL DEFAULT 0,
  `word_count` int(10) NOT NULL DEFAULT 1,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`linker_id`),
  KEY `searchwordlist_has_linker_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_linker_FKIndex2` (`linker_id`)
);

CREATE TABLE `roles` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `name` tinytext NOT NULL,
  `can_login` tinyint(1) NOT NULL DEFAULT 0,
  `is_admin` tinyint(1) NOT NULL DEFAULT 0,
  `public_profile_allowed_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `roles_name_idx` (`name`(255))
);

CREATE TABLE `user_roles` (
  `iduser_roles` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL,
  `role_id` int(10) NOT NULL,
  PRIMARY KEY (`iduser_roles`),
  KEY `user_roles_user_idx` (`users_idusers`),
  KEY `user_roles_role_idx` (`role_id`)
);

CREATE TABLE `grants` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `user_id` int(10) DEFAULT NULL,
  `role_id` int(10) DEFAULT NULL,
  `section` varchar(64) NOT NULL,
  `item` varchar(64) DEFAULT NULL,
  `rule_type` varchar(32) NOT NULL,
  `item_id` int(10) DEFAULT NULL,
  `item_rule` varchar(64) DEFAULT NULL,
  `action` varchar(64) NOT NULL,
  `extra` varchar(64) DEFAULT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
);

CREATE TABLE `preferences` (
  `idpreferences` int(10) NOT NULL AUTO_INCREMENT,
  `language_idlanguage` int(10) DEFAULT NULL,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `emailforumupdates` tinyint(1) DEFAULT 0,
  `page_size` int(10) NOT NULL DEFAULT 15,
  `auto_subscribe_replies` tinyint(1) NOT NULL DEFAULT 1,
  `timezone` tinytext DEFAULT NULL,
  PRIMARY KEY (`idpreferences`),
  KEY `preferences_FKIndex1` (`users_idusers`),
  KEY `preferences_FKIndex2` (`language_idlanguage`)
);

CREATE TABLE `searchwordlist` (
  `idsearchwordlist` int(10) NOT NULL AUTO_INCREMENT,
  `word` tinytext DEFAULT NULL,
  PRIMARY KEY (`idsearchwordlist`),
  UNIQUE KEY `searchwordlist_word_idx` (`word`(255))
);

CREATE TABLE `searchwordlist_has_linker` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `linker_id` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`linker_id`),
  KEY `searchwordlist_has_linker_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_linker_FKIndex2` (`linker_id`)
);

CREATE TABLE `site_news` (
  `idsiteNews` int(10) NOT NULL AUTO_INCREMENT,
  `forumthread_id` int(10) NOT NULL DEFAULT 0,
  `language_idlanguage` int(10) DEFAULT NULL,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `news` longtext DEFAULT NULL,
  `occurred` datetime DEFAULT NULL,
  `last_index` datetime DEFAULT NULL,
  PRIMARY KEY (`idsiteNews`),
  KEY `siteNews_FKIndex1` (`users_idusers`),
  KEY `siteNews_FKIndex2` (`language_idlanguage`),
  KEY `siteNews_FKIndex3` (`forumthread_id`)
);

CREATE TABLE `site_news_search` (
  `site_news_id` int(10) NOT NULL DEFAULT 0,
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `word_count` int(10) NOT NULL DEFAULT 1,
  PRIMARY KEY (`site_news_id`,`searchwordlist_idsearchwordlist`),
  KEY `siteNews_has_searchwordlist_FKIndex1` (`site_news_id`),
  KEY `siteNews_has_searchwordlist_FKIndex2` (`searchwordlist_idsearchwordlist`)
);


CREATE TABLE `user_language` (
  `iduserlang` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `language_idlanguage` int(10) NOT NULL DEFAULT 0,
  PRIMARY KEY (`iduserlang`),
  KEY `userpref_FKIndex1` (`language_idlanguage`),
  KEY `userpref_FKIndex2` (`users_idusers`)
);

CREATE TABLE `users` (
  `idusers` int(10) NOT NULL AUTO_INCREMENT,
  `username` tinytext DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `public_profile_enabled_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idusers`),
  UNIQUE KEY `users_username_idx` (`username`(255))
);

CREATE TABLE `passwords` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `passwd` tinytext NOT NULL,
  `passwd_algorithm` tinytext,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `passwords_user_idx` (`users_idusers`)
);

CREATE TABLE `user_emails` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `email` tinytext NOT NULL,
  `verified_at` datetime DEFAULT NULL,
  `last_verification_code` varchar(64) DEFAULT NULL,
  `verification_expires_at` datetime DEFAULT NULL,
  `notification_priority` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_emails_email_code_idx` (`email`(255),`last_verification_code`),
  KEY `user_emails_user_idx` (`user_id`)
);


CREATE TABLE `writing` (
  `idwriting` int(10) NOT NULL AUTO_INCREMENT,
  `users_idusers` int(10) NOT NULL DEFAULT 0,
  `forumthread_id` int(10) NOT NULL DEFAULT 0,
  `language_idlanguage` int(10) DEFAULT NULL,
  `writing_category_id` int(10) NOT NULL DEFAULT 0,
  `title` tinytext DEFAULT NULL,
  `published` datetime DEFAULT NULL,
  `writing` longtext DEFAULT NULL,
  `abstract` mediumtext DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `last_index` datetime DEFAULT NULL,
  PRIMARY KEY (`idwriting`),
  KEY `writing_FKIndex1` (`writing_category_id`),
  KEY `writing_FKIndex2` (`language_idlanguage`),
  KEY `writing_FKIndex3` (`forumthread_id`),
  KEY `writing_FKIndex4` (`users_idusers`)
);

CREATE TABLE `writing_category` (
  `idwritingCategory` int(10) NOT NULL AUTO_INCREMENT,
  `writing_category_id` int(10) DEFAULT NULL,
  `title` tinytext DEFAULT NULL,
  `description` tinytext DEFAULT NULL,
  PRIMARY KEY (`idwritingCategory`),
  KEY `writingCategory_FKIndex1` (`writing_category_id`)
);

CREATE TABLE `writing_search` (
  `searchwordlist_idsearchwordlist` int(10) NOT NULL DEFAULT 0,
  `writing_id` int(10) NOT NULL DEFAULT 0,
  `word_count` int(10) NOT NULL DEFAULT 1,
  PRIMARY KEY (`searchwordlist_idsearchwordlist`,`writing_id`),
  KEY `searchwordlist_has_writing_FKIndex1` (`searchwordlist_idsearchwordlist`),
  KEY `searchwordlist_has_writing_FKIndex2` (`writing_id`)
);


-- Track schema upgrades.
CREATE TABLE IF NOT EXISTS `schema_version` (
  `version` int NOT NULL
);

-- Store subscribed users for notifications.
CREATE TABLE IF NOT EXISTS `subscriptions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `method` varchar(16) NOT NULL DEFAULT 'internal',
  `pattern` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
);

-- Queue outbound emails.
CREATE TABLE IF NOT EXISTS `pending_emails` (
  `id` int NOT NULL AUTO_INCREMENT,
  `to_user_id` int DEFAULT NULL,
  `direct_email` tinyint(1) NOT NULL DEFAULT 0,
  `body` text NOT NULL,
  `error_count` int NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sent_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `pending_passwords` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `passwd` tinytext NOT NULL,
  `passwd_algorithm` tinytext NOT NULL,
  `verification_code` varchar(64) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `verified_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `pending_password_code_idx` (`verification_code`),
  KEY `pending_password_user_idx` (`user_id`)
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

-- Persist errors from asynchronous workers.
CREATE TABLE IF NOT EXISTS `dead_letters` (
  `id` int NOT NULL AUTO_INCREMENT,
  `message` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
  `path` text NOT NULL,
  `details` text,
  `data` text,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `audit_log_user_idx` (`users_idusers`)
);

CREATE TABLE IF NOT EXISTS `deactivated_users` (
  `idusers` int NOT NULL,
  `email` tinytext,
  `passwd` tinytext,
  `passwd_algorithm` tinytext,
  `username` tinytext,
  `deleted_at` datetime DEFAULT NULL,
  `restored_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idusers`)
);

CREATE TABLE IF NOT EXISTS `deactivated_comments` (
  `idcomments` int NOT NULL,
  `forumthread_id` int NOT NULL,
  `users_idusers` int NOT NULL,
  `language_idlanguage` int DEFAULT NULL,
  `written` datetime,
  `text` longtext,
  `deleted_at` datetime DEFAULT NULL,
  `restored_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idcomments`)
);

CREATE TABLE IF NOT EXISTS `deactivated_writings` (
  `idwriting` int NOT NULL,
  `users_idusers` int NOT NULL,
  `forumthread_id` int NOT NULL,
  `language_idlanguage` int DEFAULT NULL,
  `writing_category_id` int NOT NULL,
  `title` tinytext,
  `published` datetime,
  `writing` longtext,
  `abstract` mediumtext,
  `private` tinyint(1) DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `restored_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idwriting`)
);

CREATE TABLE IF NOT EXISTS `deactivated_blogs` (
  `idblogs` int NOT NULL,
  `forumthread_id` int NOT NULL,
  `users_idusers` int NOT NULL,
  `language_idlanguage` int DEFAULT NULL,
  `blog` longtext,
  `written` datetime,
  `deleted_at` datetime DEFAULT NULL,
  `restored_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idblogs`)
);

CREATE TABLE IF NOT EXISTS `deactivated_imageposts` (
  `idimagepost` int NOT NULL,
  `forumthread_id` int NOT NULL,
  `users_idusers` int NOT NULL,
  `imageboard_idimageboard` int DEFAULT NULL,
  `posted` datetime,
  `description` tinytext,
  `thumbnail` tinytext,
  `fullimage` tinytext,
  `file_size` int NOT NULL,
  `approved` tinyint(1) DEFAULT 0,
  `deleted_at` datetime DEFAULT NULL,
  `restored_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idimagepost`)
);

CREATE TABLE IF NOT EXISTS `deactivated_linker` (
  `idlinker` int NOT NULL,
  `language_idlanguage` int DEFAULT NULL,
  `users_idusers` int NOT NULL,
  `linker_category_id` int DEFAULT NULL,
  `forumthread_id` int NOT NULL,
  `title` tinytext,
  `url` tinytext,
  `description` tinytext,
  `listed` datetime,
  `deleted_at` datetime DEFAULT NULL,
  `restored_at` datetime DEFAULT NULL,
  PRIMARY KEY (`idlinker`)
);


CREATE TABLE IF NOT EXISTS `uploaded_images` (
  `iduploadedimage` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `path` tinytext,
  `width` int DEFAULT NULL,
  `height` int DEFAULT NULL,
  `file_size` int NOT NULL,
  `uploaded` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`iduploadedimage`),
  KEY `uploaded_images_user_idx` (`users_idusers`)
);

-- Comments from admins about users
CREATE TABLE IF NOT EXISTS `admin_user_comments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `comment` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `admin_user_comments_user_idx` (`users_idusers`)
);


-- Queue user requests requiring administrator action
CREATE TABLE IF NOT EXISTS `admin_request_queue` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `change_table` varchar(255) NOT NULL,
  `change_field` varchar(255) NOT NULL,
  `change_row_id` int NOT NULL,
  `change_value` text,
  `contact_options` text,
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `acted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `admin_request_queue_user_idx` (`users_idusers`)
);

-- Comments for administrator requests
CREATE TABLE IF NOT EXISTS `admin_request_comments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `request_id` int NOT NULL,
  `comment` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `admin_request_comments_request_idx` (`request_id`)
);

CREATE TABLE IF NOT EXISTS `external_links` (
  `id` int NOT NULL AUTO_INCREMENT,
  `url` tinytext NOT NULL,
  `clicks` int NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` int DEFAULT NULL,
  `card_title` tinytext,
  `card_description` tinytext,
  `card_image` tinytext,
  `card_image_cache` tinytext,
  `favicon_cache` tinytext,
  PRIMARY KEY (`id`),
  UNIQUE KEY `external_links_url_idx` (`url`(255))
);
