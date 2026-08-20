PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE goose_db_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp TIMESTAMP DEFAULT (datetime('now'))
	);
INSERT INTO goose_db_version VALUES(1,0,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(2,1,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(3,2,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(4,3,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(5,4,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(6,5,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(7,6,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(8,7,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(9,8,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(10,9,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(11,10,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(12,11,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(13,12,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(14,13,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(15,14,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(16,15,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(17,16,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(18,17,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(19,18,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(20,19,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(21,20,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(22,21,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(23,22,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(24,23,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(25,24,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(26,25,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(27,26,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(28,27,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(29,28,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(30,29,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(31,30,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(32,31,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(33,32,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(34,33,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(35,34,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(36,35,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(37,36,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(38,37,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(39,38,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(40,39,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(41,40,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(42,41,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(43,42,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(44,43,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(45,44,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(46,45,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(47,46,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(48,47,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(49,48,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(50,49,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(51,50,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(52,51,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(53,52,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(54,53,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(55,54,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(56,55,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(57,56,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(58,57,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(59,58,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(60,59,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(61,60,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(62,61,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(63,62,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(64,63,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(65,64,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(66,65,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(67,66,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(68,67,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(69,68,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(70,69,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(71,70,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(72,71,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(73,72,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(74,73,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(75,74,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(76,75,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(77,76,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(78,77,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(79,78,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(80,79,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(81,80,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(82,81,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(83,82,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(84,83,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(85,84,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(86,85,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(87,86,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(88,87,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(89,88,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(90,89,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(91,90,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(92,91,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(93,92,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(94,93,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(95,94,1,'2026-08-19 13:40:45');
INSERT INTO goose_db_version VALUES(96,95,1,'2026-08-19 13:40:45');
CREATE TABLE `1_old_forumthread` (
idforumthread INTEGER PRIMARY KEY AUTOINCREMENT,
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE `1_old_forumtopic` (
idforumtopic INTEGER PRIMARY KEY AUTOINCREMENT,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);
CREATE TABLE blogs (
idblogs INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0,
blog TEXT DEFAULT NULL,
written DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
, deleted_at DATETIME DEFAULT NULL, last_index datetime DEFAULT NULL, timezone TEXT DEFAULT NULL);
INSERT INTO blogs VALUES(1,0,1,1,'Welcome to the official developer blog for Goa4Web.','2026-08-19 13:40:45',NULL,NULL,NULL);
CREATE TABLE IF NOT EXISTS "blogs_search" (
blog_id INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0, word_count INT NOT NULL DEFAULT 1,
PRIMARY KEY (blog_id,searchwordlist_idsearchwordlist)
);
CREATE TABLE bookmarks (
idbookmarks INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
list BLOB DEFAULT NULL
);
INSERT INTO bookmarks VALUES(1,2,'{"thread_ids":[1,2]}');
CREATE TABLE comments (
idcomments INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0,
written DATETIME DEFAULT NULL,
TEXT TEXT DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, last_index datetime DEFAULT NULL, timezone TEXT DEFAULT NULL);
INSERT INTO comments VALUES(1,1,1,1,'2026-08-19 13:40:45','Welcome to the Goa4Web community platform!',NULL,NULL,NULL);
INSERT INTO comments VALUES(2,1,2,1,'2026-08-19 13:40:45','Thanks! Excited to be part of the community.',NULL,NULL,NULL);
INSERT INTO comments VALUES(3,2,2,1,'2026-08-19 13:40:45','How do I run Goa4Web with the SQLite backend?',NULL,NULL,NULL);
INSERT INTO comments VALUES(4,3,1,1,'2026-08-19 13:40:45','Confidential staff notes for internal review.',NULL,NULL,NULL);
CREATE TABLE IF NOT EXISTS "comments_search" (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
comment_id INTEGER NOT NULL DEFAULT 0, word_count INT NOT NULL DEFAULT 1,
PRIMARY KEY (searchwordlist_idsearchwordlist,comment_id)
);
CREATE TABLE faq (
id INTEGER PRIMARY KEY AUTOINCREMENT,
category_id INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0,
author_id INTEGER NOT NULL DEFAULT 0,
answer TEXT DEFAULT NULL,
question TEXT DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, priority INT NOT NULL DEFAULT 0, updated_at DATETIME DEFAULT NULL, description TEXT DEFAULT '');
INSERT INTO faq VALUES(1,1,1,1,'Goa4Web is an open source community platform written in Go.','What is Goa4Web?',NULL,1,NULL,'Overview FAQ');
CREATE TABLE IF NOT EXISTS "faq_categories" (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, parent_category_id INT NULL DEFAULT NULL, language_id INT NULL DEFAULT NULL, updated_at DATETIME DEFAULT NULL, priority INT NOT NULL DEFAULT 0);
INSERT INTO faq_categories VALUES(1,'General FAQ',NULL,NULL,1,NULL,1);
CREATE TABLE forumcategory (
idforumcategory INTEGER PRIMARY KEY AUTOINCREMENT,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, language_id INT NOT NULL DEFAULT 0);
INSERT INTO forumcategory VALUES(1,0,'General Discussion','Community discussions and chatter',NULL,1);
INSERT INTO forumcategory VALUES(2,0,'Technology','Technical questions, software, and programming',NULL,1);
CREATE TABLE forumthread (
idforumthread INTEGER PRIMARY KEY AUTOINCREMENT,
firstpost INTEGER NOT NULL DEFAULT 0,
lastposter INTEGER NOT NULL DEFAULT 0,
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
comments INTEGER DEFAULT NULL,
lastaddition DATETIME DEFAULT NULL,
locked INTEGER DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, reply_to_comment_id INTEGER DEFAULT NULL, reply_to_thread_id INTEGER DEFAULT NULL);
INSERT INTO forumthread VALUES(1,1,2,1,2,'2026-08-19 13:40:45',0,NULL,NULL,NULL);
INSERT INTO forumthread VALUES(2,3,2,3,1,'2026-08-19 13:40:45',0,NULL,NULL,NULL);
INSERT INTO forumthread VALUES(3,4,1,4,1,'2026-08-19 13:40:45',0,NULL,NULL,NULL);
CREATE TABLE forumtopic (
idforumtopic INTEGER PRIMARY KEY AUTOINCREMENT,
lastposter INTEGER NOT NULL DEFAULT 0,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
threads INTEGER DEFAULT NULL,
comments INTEGER DEFAULT NULL,
lastaddition DATETIME DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, language_id INT NOT NULL DEFAULT 0, handler TEXT NOT NULL DEFAULT '');
INSERT INTO forumtopic VALUES(1,1,1,'Welcome & Rules','Site announcements and guidelines',1,2,NULL,NULL,1,'');
INSERT INTO forumtopic VALUES(2,1,1,'General Chit-Chat','Chat about anything',0,0,NULL,NULL,1,'');
INSERT INTO forumtopic VALUES(3,2,2,'Go Programming','Discussions about the Go language and Goa4Web',1,1,NULL,NULL,1,'');
INSERT INTO forumtopic VALUES(4,1,0,'Staff Room','Private forum for administrators and moderators',1,1,NULL,NULL,1,'private');
CREATE TABLE imageboard (
idimageboard INTEGER PRIMARY KEY AUTOINCREMENT,
imageboard_idimageboard INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
, approval_required TINYINT(1) NOT NULL DEFAULT 0, deleted_at DATETIME DEFAULT NULL);
CREATE TABLE imagepost (
idimagepost INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
imageboard_idimageboard INTEGER NOT NULL DEFAULT 0,
posted DATETIME DEFAULT NULL,
description TEXT DEFAULT NULL,
thumbnail TEXT DEFAULT NULL,
fullimage TEXT DEFAULT NULL
, approved TINYINT(1) NOT NULL DEFAULT 0, file_size INT NOT NULL DEFAULT 0, deleted_at DATETIME DEFAULT NULL, last_index datetime DEFAULT NULL, timezone TEXT DEFAULT NULL);
CREATE TABLE IF NOT EXISTS "imagepost_search" (
image_post_id INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0, word_count INT NOT NULL DEFAULT 1,
PRIMARY KEY (image_post_id,searchwordlist_idsearchwordlist)
);
CREATE TABLE language (
id INTEGER PRIMARY KEY AUTOINCREMENT,
nameof TEXT DEFAULT NULL
);
INSERT INTO language VALUES(1,'English');
CREATE TABLE linker (
id INTEGER PRIMARY KEY AUTOINCREMENT,
language_id INTEGER NOT NULL DEFAULT 0,
author_id INTEGER NOT NULL DEFAULT 0,
category_id INTEGER NOT NULL DEFAULT 0,
thread_id INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
url TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
listed DATETIME DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, last_index datetime DEFAULT NULL, timezone TEXT DEFAULT NULL);
INSERT INTO linker VALUES(1,1,1,1,0,'Official Go Website','https://go.dev','The home of the Go programming language.','2026-08-19 13:40:45',NULL,NULL,NULL);
CREATE TABLE IF NOT EXISTS "linker_category" (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT DEFAULT NULL
, position INT NOT NULL DEFAULT 0, sortorder INT NOT NULL DEFAULT 0);
INSERT INTO linker_category VALUES(1,'Useful Resources',1,1);
CREATE TABLE IF NOT EXISTS "linker_queue" (
id INTEGER PRIMARY KEY AUTOINCREMENT,
language_id INTEGER NOT NULL DEFAULT 0,
submitter_id INTEGER NOT NULL DEFAULT 0,
category_id INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
url TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
, timezone TEXT DEFAULT NULL);
CREATE TABLE IF NOT EXISTS "linker_search" (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
linker_id INTEGER NOT NULL DEFAULT 0, word_count INT NOT NULL DEFAULT 1,
PRIMARY KEY (searchwordlist_idsearchwordlist,linker_id)
);
CREATE TABLE IF NOT EXISTS "user_roles" (
iduser_roles INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
section TEXT DEFAULT NULL,
role tinyblob DEFAULT NULL
, role_id INT);
INSERT INTO user_roles VALUES(1,1,NULL,NULL,6);
INSERT INTO user_roles VALUES(2,2,NULL,NULL,4);
INSERT INTO user_roles VALUES(3,3,NULL,NULL,1);
CREATE TABLE preferences (
idpreferences INTEGER PRIMARY KEY AUTOINCREMENT,
language_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
emailforumupdates INTEGER DEFAULT 0
, page_size INT NOT NULL DEFAULT 15, auto_subscribe_replies TINYINT(1) NOT NULL DEFAULT 1, timezone TEXT DEFAULT NULL, custom_css TEXT, daily_digest_hour INT DEFAULT NULL, daily_digest_mark_read TINYINT(1) NOT NULL DEFAULT 0, last_digest_sent_at DATETIME DEFAULT NULL, weekly_digest_day INT DEFAULT NULL, weekly_digest_hour INT DEFAULT NULL, last_weekly_digest_sent_at DATETIME DEFAULT NULL, monthly_digest_day INT DEFAULT NULL, monthly_digest_hour INT DEFAULT NULL, last_monthly_digest_sent_at DATETIME DEFAULT NULL, image_safe_dimension TEXT);
INSERT INTO preferences VALUES(1,1,1,1,15,1,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO preferences VALUES(2,1,2,1,15,1,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO preferences VALUES(3,1,3,1,15,1,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE searchwordlist (
idsearchwordlist INTEGER PRIMARY KEY AUTOINCREMENT,
word TEXT DEFAULT NULL
);
CREATE TABLE searchwordlist_has_linker (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
linker_id INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (searchwordlist_idsearchwordlist,linker_id)
);
CREATE TABLE IF NOT EXISTS "site_news" (
idsiteNews INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
news TEXT DEFAULT NULL,
occurred DATETIME DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, last_index datetime DEFAULT NULL, timezone TEXT DEFAULT NULL);
INSERT INTO site_news VALUES(1,0,1,1,'Goa4Web now supports SQLite backend out of the box!','2026-08-19 13:40:45',NULL,NULL,NULL);
CREATE TABLE IF NOT EXISTS "site_news_search" (
site_news_id INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0, word_count INT NOT NULL DEFAULT 1,
PRIMARY KEY (site_news_id,searchwordlist_idsearchwordlist)
);
CREATE TABLE IF NOT EXISTS "topic_permissions" (
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
view_role_id INTEGER DEFAULT NULL,
reply_role_id INTEGER DEFAULT NULL,
newthread_role_id INTEGER DEFAULT NULL,
see_role_id INTEGER DEFAULT NULL,
invite_role_id INTEGER DEFAULT NULL,
read_role_id INTEGER DEFAULT NULL,
mod_role_id INTEGER DEFAULT NULL,
admin_role_id INTEGER DEFAULT NULL
);
CREATE TABLE IF NOT EXISTS "user_language" (
iduserlang INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0
);
INSERT INTO user_language VALUES(1,1,1);
INSERT INTO user_language VALUES(2,2,1);
INSERT INTO user_language VALUES(3,3,1);
CREATE TABLE users (
idusers INTEGER PRIMARY KEY AUTOINCREMENT,
email TEXT DEFAULT NULL,
passwd TEXT DEFAULT NULL,
username TEXT DEFAULT NULL
, passwd_algorithm TEXT DEFAULT NULL, deleted_at DATETIME DEFAULT NULL, public_profile_enabled_at DATETIME DEFAULT NULL);
INSERT INTO users VALUES(1,'admin@example.com',NULL,'admin',NULL,NULL,NULL);
INSERT INTO users VALUES(2,'user@example.com',NULL,'testuser',NULL,NULL,NULL);
INSERT INTO users VALUES(3,'writer@example.com',NULL,'writer',NULL,NULL,NULL);
CREATE TABLE IF NOT EXISTS "user_topic_permissions" (
users_idusers INTEGER NOT NULL DEFAULT 0,
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
role_id INTEGER DEFAULT NULL,
invitemax INTEGER DEFAULT NULL, expires_at DATETIME DEFAULT NULL,
PRIMARY KEY (users_idusers,forumtopic_idforumtopic)
);
CREATE TABLE writing (
idwriting INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
forumthread_id INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0,
writing_category_id INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
published DATETIME DEFAULT NULL,
writing TEXT DEFAULT NULL,
abstract TEXT DEFAULT NULL,
private INTEGER DEFAULT NULL
, deleted_at DATETIME DEFAULT NULL, last_index datetime DEFAULT NULL, timezone TEXT DEFAULT NULL);
CREATE TABLE IF NOT EXISTS "writing_category" (
idwritingCategory INTEGER PRIMARY KEY AUTOINCREMENT,
writing_category_id INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);
CREATE TABLE IF NOT EXISTS "writing_search" (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
writing_id INTEGER NOT NULL DEFAULT 0, word_count INT NOT NULL DEFAULT 1,
PRIMARY KEY (searchwordlist_idsearchwordlist,writing_id)
);
CREATE TABLE schema_version (
version int NOT NULL
);
INSERT INTO schema_version VALUES(95);
CREATE TABLE subscriptions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
item_type TEXT NOT NULL,
target_id int NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
, method TEXT NOT NULL DEFAULT 'internal', pattern TEXT NOT NULL DEFAULT '');
CREATE TABLE pending_emails (
id INTEGER PRIMARY KEY AUTOINCREMENT,
to_email TEXT NOT NULL,
subject TEXT NOT NULL,
body TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
sent_at DATETIME DEFAULT NULL
, html_body TEXT, error_count INT NOT NULL DEFAULT 0, to_user_id INT NOT NULL DEFAULT 0, direct_email TINYINT(1) NOT NULL DEFAULT 0);
CREATE TABLE notifications (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
link TEXT,
message TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
read_at DATETIME DEFAULT NULL
);
CREATE TABLE sessions (
session_id TEXT NOT NULL,
users_idusers int NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (session_id)
);
CREATE TABLE site_announcements (
id INTEGER PRIMARY KEY AUTOINCREMENT,
site_news_id int NOT NULL,
active INTEGER NOT NULL DEFAULT 1,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE login_attempts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL,
ip_address TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE banned_ips (
id INTEGER PRIMARY KEY AUTOINCREMENT,
ip_net TEXT NOT NULL,
reason TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
expires_at DATETIME DEFAULT NULL,
canceled_at DATETIME DEFAULT NULL,
UNIQUE (ip_net)
);
CREATE TABLE template_overrides (
name TEXT NOT NULL,
body TEXT NOT NULL,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
PRIMARY KEY (name)
);
CREATE TABLE audit_log (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
action TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
, path TEXT NOT NULL DEFAULT '', details TEXT, data TEXT);
CREATE TABLE deactivated_users (
idusers INT NOT NULL,
email TEXT,
passwd TEXT,
passwd_algorithm TEXT,
username TEXT,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idusers)
);
CREATE TABLE deactivated_comments (
idcomments INT NOT NULL,
forumthread_id INT NOT NULL,
users_idusers INT NOT NULL,
language_id INT NOT NULL,
written DATETIME,
TEXT TEXT,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL, timezone TEXT DEFAULT NULL,
PRIMARY KEY (idcomments)
);
CREATE TABLE deactivated_writings (
idwriting INT NOT NULL,
users_idusers INT NOT NULL,
forumthread_id INT NOT NULL,
language_id INT NOT NULL,
writingCategory_idwritingCategory INT NOT NULL,
title TEXT,
published DATETIME,
writing TEXT,
abstract TEXT,
private INTEGER DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL, timezone TEXT DEFAULT NULL,
PRIMARY KEY (idwriting)
);
CREATE TABLE deactivated_blogs (
idblogs INT NOT NULL,
forumthread_id INT NOT NULL,
users_idusers INT NOT NULL,
language_id INT NOT NULL,
blog TEXT,
written DATETIME,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL, timezone TEXT DEFAULT NULL,
PRIMARY KEY (idblogs)
);
CREATE TABLE deactivated_imageposts (
idimagepost INT NOT NULL,
forumthread_id INT NOT NULL,
users_idusers INT NOT NULL,
imageboard_idimageboard INT NOT NULL,
posted DATETIME,
description TEXT,
thumbnail TEXT,
fullimage TEXT,
file_size INT NOT NULL,
approved INTEGER DEFAULT 0,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL, timezone TEXT DEFAULT NULL,
PRIMARY KEY (idimagepost)
);
CREATE TABLE deactivated_linker (
id INT NOT NULL,
language_id INT NOT NULL,
author_id INT NOT NULL,
category_id INT NOT NULL,
thread_id INT NOT NULL,
title TEXT,
url TEXT,
description TEXT,
listed DATETIME,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL, timezone TEXT DEFAULT NULL,
PRIMARY KEY (id)
);
CREATE TABLE IF NOT EXISTS "dead_letters" (
id INTEGER PRIMARY KEY AUTOINCREMENT,
message TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO passwords VALUES(1,1,'password123','plaintext','2026-08-19 13:40:45');
INSERT INTO passwords VALUES(2,2,'password123','plaintext','2026-08-19 13:40:45');
INSERT INTO passwords VALUES(3,3,'password123','plaintext','2026-08-19 13:40:45');
CREATE TABLE uploaded_images (
iduploadedimage INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
path TEXT,
thumbnail TEXT,
file_size INT NOT NULL,
uploaded DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
, width INT DEFAULT NULL, height INT DEFAULT NULL);
CREATE TABLE user_emails (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id int NOT NULL,
email TEXT NOT NULL,
verified INTEGER NOT NULL DEFAULT 0,
last_verification_code TEXT DEFAULT NULL, verified_at datetime DEFAULT NULL, notification_priority int NOT NULL DEFAULT 0, verification_expires_at datetime DEFAULT NULL,
UNIQUE (user_id, email)
);
INSERT INTO user_emails VALUES(1,1,'admin@example.com',0,NULL,'2026-08-19 13:40:45',100,NULL);
INSERT INTO user_emails VALUES(2,2,'user@example.com',0,NULL,'2026-08-19 13:40:45',100,NULL);
INSERT INTO user_emails VALUES(3,3,'writer@example.com',0,NULL,'2026-08-19 13:40:45',100,NULL);
CREATE TABLE pending_passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id int NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT,
verification_code TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
verified_at DATETIME DEFAULT NULL,
UNIQUE (verification_code)
);
CREATE TABLE roles (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL, can_login TINYINT(1) NOT NULL DEFAULT 0, is_admin TINYINT(1) NOT NULL DEFAULT 0, public_profile_allowed_at DATETIME DEFAULT NULL, private_labels TINYINT(1) NOT NULL DEFAULT 1,
UNIQUE (name)
);
INSERT INTO roles VALUES(1,'content writer',1,0,'2026-08-19 13:40:45',1);
INSERT INTO roles VALUES(2,'rejected',0,0,NULL,0);
INSERT INTO roles VALUES(3,'anyone',0,0,NULL,0);
INSERT INTO roles VALUES(4,'user',1,0,NULL,1);
INSERT INTO roles VALUES(6,'administrator',1,1,NULL,1);
CREATE TABLE grants (
id INTEGER PRIMARY KEY AUTOINCREMENT,
created_at DATETIME NULL,
updated_at DATETIME NULL,
user_id INT NULL,
role_id INT NULL,
section TEXT NOT NULL,
item TEXT NULL,
rule_type TEXT NOT NULL,
item_id INT NULL,
item_rule TEXT NULL,
action TEXT NOT NULL,
extra TEXT NULL,
active INTEGER NOT NULL DEFAULT 1
);
INSERT INTO grants VALUES(1,'2026-08-19 13:40:45',NULL,NULL,1,'news','post','allow',NULL,NULL,'post',NULL,1);
INSERT INTO grants VALUES(2,'2026-08-19 13:40:45',NULL,NULL,1,'linker','link','allow',NULL,NULL,'post',NULL,1);
INSERT INTO grants VALUES(3,'2026-08-19 13:40:45',NULL,NULL,1,'search',NULL,'allow',NULL,NULL,'search',NULL,1);
INSERT INTO grants VALUES(4,'2026-08-19 13:40:45',NULL,NULL,1,'news',NULL,'allow',NULL,NULL,'search',NULL,1);
INSERT INTO grants VALUES(5,'2026-08-19 13:40:45',NULL,NULL,1,'forum',NULL,'allow',NULL,NULL,'search',NULL,1);
INSERT INTO grants VALUES(6,'2026-08-19 13:40:45',NULL,NULL,1,'linker',NULL,'allow',NULL,NULL,'search',NULL,1);
INSERT INTO grants VALUES(7,'2026-08-19 13:40:45',NULL,NULL,1,'blogs',NULL,'allow',NULL,NULL,'search',NULL,1);
INSERT INTO grants VALUES(8,'2026-08-19 13:40:45',NULL,NULL,1,'writing',NULL,'allow',NULL,NULL,'search',NULL,1);
INSERT INTO grants VALUES(9,'2026-08-19 13:40:45',NULL,NULL,1,'images','upload','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(10,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum','topic','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(11,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum','topic','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(12,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum','topic','allow',NULL,NULL,'reply',NULL,1);
INSERT INTO grants VALUES(13,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum','topic','allow',NULL,NULL,'post',NULL,1);
INSERT INTO grants VALUES(14,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum','topic','allow',NULL,NULL,'edit',NULL,1);
INSERT INTO grants VALUES(15,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum','topic','allow',NULL,NULL,'create',NULL,1);
INSERT INTO grants VALUES(16,'2026-08-19 13:40:45',NULL,NULL,NULL,'blogs','entry','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(17,'2026-08-19 13:40:45',NULL,NULL,NULL,'blogs','entry','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(18,'2026-08-19 13:40:45',NULL,NULL,NULL,'writing','category','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(19,'2026-08-19 13:40:45',NULL,NULL,NULL,'writing','category','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(20,'2026-08-19 13:40:45',NULL,NULL,NULL,'writing','article','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(21,'2026-08-19 13:40:45',NULL,NULL,NULL,'writing','article','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(22,'2026-08-19 13:40:45',NULL,NULL,NULL,'news','post','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(23,'2026-08-19 13:40:45',NULL,NULL,NULL,'news','post','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(24,'2026-08-19 13:40:45',NULL,NULL,NULL,'faq','category','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(25,'2026-08-19 13:40:45',NULL,NULL,NULL,'faq','category','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(26,'2026-08-19 13:40:45',NULL,NULL,NULL,'faq','question/answer','allow',NULL,NULL,'see',NULL,1);
INSERT INTO grants VALUES(27,'2026-08-19 13:40:45',NULL,NULL,NULL,'faq','question/answer','allow',NULL,NULL,'view',NULL,1);
INSERT INTO grants VALUES(28,'2026-08-19 13:40:45',NULL,NULL,1,'images',NULL,'allow',NULL,NULL,'label',NULL,1);
INSERT INTO grants VALUES(29,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum',NULL,'allow',NULL,NULL,'label',NULL,1);
INSERT INTO grants VALUES(30,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum',NULL,'allow',NULL,NULL,'label',NULL,1);
INSERT INTO grants VALUES(31,'2026-08-19 13:40:45',NULL,NULL,1,'images',NULL,'allow',NULL,NULL,'label',NULL,1);
INSERT INTO grants VALUES(32,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum',NULL,'allow',NULL,NULL,'label',NULL,1);
INSERT INTO grants VALUES(33,'2026-08-19 13:40:45',NULL,NULL,1,'privateforum',NULL,'allow',NULL,NULL,'label',NULL,1);
INSERT INTO grants VALUES(34,'2026-08-19 13:40:45',NULL,NULL,6,'role',NULL,'allow',NULL,NULL,'moderator',NULL,1);
INSERT INTO grants VALUES(35,'2026-08-19 13:40:45',NULL,NULL,6,'role',NULL,'allow',NULL,NULL,'content writer',NULL,1);
INSERT INTO grants VALUES(36,'2026-08-19 13:40:45',NULL,1,NULL,'privateforum','topic','allow',4,NULL,'see',NULL,1);
INSERT INTO grants VALUES(37,'2026-08-19 13:40:45',NULL,1,NULL,'privateforum','topic','allow',4,NULL,'view',NULL,1);
INSERT INTO grants VALUES(38,'2026-08-19 13:40:45',NULL,1,NULL,'privateforum','topic','allow',4,NULL,'post',NULL,1);
INSERT INTO grants VALUES(39,'2026-08-19 13:40:45',NULL,1,NULL,'privateforum','topic','allow',4,NULL,'reply',NULL,1);
INSERT INTO grants VALUES(40,'2026-08-19 13:40:45',NULL,1,NULL,'privateforum_thread','thread','allow',3,NULL,'view',NULL,1);
INSERT INTO grants VALUES(41,'2026-08-19 13:40:45',NULL,1,NULL,'privateforum_thread','thread','allow',3,NULL,'reply',NULL,1);
CREATE TABLE admin_user_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE admin_request_queue (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
change_table TEXT NOT NULL,
change_field TEXT NOT NULL,
change_row_id int NOT NULL,
change_value TEXT,
contact_options TEXT,
status TEXT NOT NULL DEFAULT 'pending',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
acted_at DATETIME DEFAULT NULL
);
CREATE TABLE admin_request_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
request_id INT NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE faq_revisions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
faq_id INT NOT NULL,
users_idusers INT NOT NULL,
question TEXT,
answer TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
, timezone TEXT DEFAULT NULL);
CREATE TABLE external_links (
id INTEGER PRIMARY KEY AUTOINCREMENT,
url TEXT NOT NULL,
clicks INT NOT NULL DEFAULT 0,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
updated_by INT DEFAULT NULL,
card_title TEXT,
card_description TEXT,
card_image TEXT,
card_image_cache TEXT,
favicon_cache TEXT, card_duration TEXT, card_upload_date TEXT, card_author TEXT,
UNIQUE (url)
);
CREATE TABLE IF NOT EXISTS "content_public_labels" (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item_id INT NOT NULL,
label TEXT NOT NULL, item TEXT NOT NULL DEFAULT 'thread',
UNIQUE (item_id, label)
);
CREATE TABLE IF NOT EXISTS "content_private_labels" (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item_id INT NOT NULL,
user_id INT NOT NULL,
label TEXT NOT NULL,
invert INTEGER NOT NULL DEFAULT 0, item TEXT NOT NULL DEFAULT 'thread',
UNIQUE (item_id, user_id, label)
);
CREATE TABLE content_label_status (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id INT NOT NULL,
label TEXT NOT NULL,
UNIQUE (item, item_id, label)
);
CREATE TABLE content_read_markers (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id INT NOT NULL,
user_id INT NOT NULL,
last_comment_id INT NOT NULL,
UNIQUE (item, item_id, user_id)
);
CREATE TABLE role_subscription_archetypes (
id INTEGER PRIMARY KEY AUTOINCREMENT,
role_id int NOT NULL,
archetype_name TEXT NOT NULL,
pattern TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE thread_images (
idthread_image INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INT NOT NULL,
path TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE scheduler_state (
task_name TEXT NOT NULL PRIMARY KEY,
last_run_at DATETIME DEFAULT NULL,
metadata TEXT DEFAULT NULL
);
CREATE TABLE image_cache_entries (
id TEXT NOT NULL,
source_url TEXT DEFAULT NULL,
source_kind TEXT NOT NULL DEFAULT 'unknown',
status TEXT NOT NULL DEFAULT 'ready',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
last_used_at DATETIME DEFAULT NULL,
fetched_at DATETIME DEFAULT NULL,
expires_at DATETIME DEFAULT NULL,
content_expires_at DATETIME DEFAULT NULL,
content_type TEXT DEFAULT NULL,
size_bytes bigint DEFAULT NULL,
width int DEFAULT NULL,
height int DEFAULT NULL,
checksum TEXT DEFAULT NULL,
thumbnail_id TEXT DEFAULT NULL,
error_message TEXT DEFAULT NULL,
retry_count int NOT NULL DEFAULT 0,
last_attempt_at DATETIME DEFAULT NULL,
next_attempt_at DATETIME DEFAULT NULL, uploaded_image_id INT DEFAULT NULL,
PRIMARY KEY (id)
);
CREATE TABLE api_keys (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
api_key TEXT NOT NULL,
name TEXT NOT NULL,
scopes TEXT NOT NULL,
expires_at DATETIME DEFAULT NULL,
last_used_at DATETIME DEFAULT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
revoked_at DATETIME DEFAULT NULL,
UNIQUE (api_key)
);
CREATE TABLE user_passkeys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INT NOT NULL,
    credential_id BLOB NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type TEXT NOT NULL,
    aaguid BLOB NOT NULL,
    sign_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL, name TEXT NOT NULL DEFAULT 'Passkey', backup_eligible BOOLEAN DEFAULT NULL, backup_state BOOLEAN DEFAULT NULL,
    UNIQUE (credential_id)
);
PRAGMA writable_schema=ON;
CREATE TABLE IF NOT EXISTS sqlite_sequence(name,seq);
DELETE FROM sqlite_sequence;
INSERT INTO sqlite_sequence VALUES('goose_db_version',96);
INSERT INTO sqlite_sequence VALUES('user_emails',3);
INSERT INTO sqlite_sequence VALUES('roles',6);
INSERT INTO sqlite_sequence VALUES('grants',41);
INSERT INTO sqlite_sequence VALUES('user_roles',3);
INSERT INTO sqlite_sequence VALUES('language',1);
INSERT INTO sqlite_sequence VALUES('users',3);
INSERT INTO sqlite_sequence VALUES('passwords',3);
INSERT INTO sqlite_sequence VALUES('user_language',3);
INSERT INTO sqlite_sequence VALUES('preferences',3);
INSERT INTO sqlite_sequence VALUES('forumcategory',2);
INSERT INTO sqlite_sequence VALUES('forumtopic',4);
INSERT INTO sqlite_sequence VALUES('forumthread',3);
INSERT INTO sqlite_sequence VALUES('comments',4);
INSERT INTO sqlite_sequence VALUES('site_news',1);
INSERT INTO sqlite_sequence VALUES('blogs',1);
INSERT INTO sqlite_sequence VALUES('linker_category',1);
INSERT INTO sqlite_sequence VALUES('linker',1);
INSERT INTO sqlite_sequence VALUES('faq_categories',1);
INSERT INTO sqlite_sequence VALUES('faq',1);
INSERT INTO sqlite_sequence VALUES('bookmarks',1);
CREATE UNIQUE INDEX users_username_idx ON users (username);
CREATE UNIQUE INDEX users_email_idx ON users (email);
CREATE UNIQUE INDEX topicrestrictions_topic_idx ON "topic_permissions" (forumtopic_idforumtopic);
CREATE UNIQUE INDEX user_emails_email_idx ON user_emails (email);
CREATE UNIQUE INDEX topicrestrictions_forumtopic_idx ON "topic_permissions" (forumtopic_idforumtopic);
CREATE UNIQUE INDEX user_emails_email_code_idx ON user_emails (email, last_verification_code);
CREATE INDEX forumcategory_FKIndex2 ON forumcategory (language_id);
CREATE INDEX forumtopic_FKIndex3 ON forumtopic (language_id);
CREATE UNIQUE INDEX content_public_labels_uq ON content_public_labels (item, item_id, label);
CREATE UNIQUE INDEX content_private_labels_uq ON content_private_labels (item, item_id, user_id, label);
CREATE INDEX faq_priority_idx ON faq (priority);
CREATE INDEX image_cache_entries_uploaded_image_idx ON image_cache_entries (uploaded_image_id);
CREATE INDEX forumthread_reply_to_thread_id ON forumthread(reply_to_thread_id, reply_to_comment_id);
PRAGMA writable_schema=OFF;
COMMIT;
