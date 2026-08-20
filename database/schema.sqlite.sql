CREATE TABLE blogs (
idblogs INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER DEFAULT NULL,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_id INTEGER DEFAULT NULL,
blog TEXT DEFAULT NULL,
written DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
timezone TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
last_index DATETIME DEFAULT NULL
);

CREATE TABLE blogs_search (
blog_id INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
word_count INTEGER NOT NULL DEFAULT 1,
PRIMARY KEY (blog_id,searchwordlist_idsearchwordlist)
);

CREATE TABLE bookmarks (
idbookmarks INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
list BLOB DEFAULT NULL
);

CREATE TABLE comments (
idcomments INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_id INTEGER DEFAULT NULL,
written DATETIME DEFAULT NULL,
TEXT TEXT DEFAULT NULL,
timezone TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
last_index DATETIME DEFAULT NULL
);

CREATE TABLE comments_search (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
comment_id INTEGER NOT NULL DEFAULT 0,
word_count INTEGER NOT NULL DEFAULT 1,
PRIMARY KEY (searchwordlist_idsearchwordlist,comment_id)
);

CREATE TABLE faq (
id INTEGER PRIMARY KEY AUTOINCREMENT,
category_id INTEGER DEFAULT NULL,
language_id INTEGER DEFAULT NULL,
author_id INTEGER NOT NULL DEFAULT 0,
answer TEXT DEFAULT NULL,
question TEXT DEFAULT NULL,
description TEXT DEFAULT '',
priority INT NOT NULL DEFAULT 0,
deleted_at DATETIME DEFAULT NULL,
updated_at DATETIME DEFAULT NULL
);

CREATE TABLE faq_categories (
id INTEGER PRIMARY KEY AUTOINCREMENT,
parent_category_id INTEGER DEFAULT NULL,
language_id INTEGER DEFAULT NULL,
name TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
priority INT NOT NULL DEFAULT 0,
updated_at DATETIME DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS faq_revisions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
faq_id int NOT NULL,
users_idusers int NOT NULL,
question TEXT,
answer TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
timezone TEXT DEFAULT NULL
);

CREATE TABLE forumcategory (
idforumcategory INTEGER PRIMARY KEY AUTOINCREMENT,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
language_id INTEGER DEFAULT NULL,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL
);

CREATE TABLE forumthread (
idforumthread INTEGER PRIMARY KEY AUTOINCREMENT,
firstpost INTEGER NOT NULL DEFAULT 0,
lastposter INTEGER NOT NULL DEFAULT 0,
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
comments INTEGER DEFAULT NULL,
lastaddition DATETIME DEFAULT NULL,
locked INTEGER DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
reply_to_comment_id INTEGER DEFAULT NULL,
reply_to_thread_id INTEGER DEFAULT NULL
);

CREATE TABLE forumtopic (
idforumtopic INTEGER PRIMARY KEY AUTOINCREMENT,
lastposter INTEGER NOT NULL DEFAULT 0,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
language_id INTEGER DEFAULT NULL,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
threads INTEGER DEFAULT NULL,
comments INTEGER DEFAULT NULL,
lastaddition DATETIME DEFAULT NULL,
handler TEXT NOT NULL DEFAULT '',
deleted_at DATETIME DEFAULT NULL
);

CREATE TABLE imageboard (
idimageboard INTEGER PRIMARY KEY AUTOINCREMENT,
imageboard_idimageboard INTEGER DEFAULT NULL,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
approval_required INTEGER NOT NULL DEFAULT 0,
deleted_at DATETIME DEFAULT NULL
);

CREATE TABLE imagepost (
idimagepost INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
imageboard_idimageboard INTEGER DEFAULT NULL,
posted DATETIME DEFAULT NULL,
timezone TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
thumbnail TEXT DEFAULT NULL,
fullimage TEXT DEFAULT NULL,
file_size INTEGER NOT NULL DEFAULT 0,
approved INTEGER NOT NULL DEFAULT 0,
deleted_at DATETIME DEFAULT NULL,
last_index DATETIME DEFAULT NULL
);

CREATE TABLE imagepost_search (
image_post_id INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
word_count INTEGER NOT NULL DEFAULT 1,
PRIMARY KEY (image_post_id,searchwordlist_idsearchwordlist)
);

CREATE TABLE language (
id INTEGER PRIMARY KEY AUTOINCREMENT,
nameof TEXT DEFAULT NULL
);

CREATE TABLE linker (
id INTEGER PRIMARY KEY AUTOINCREMENT,
language_id INTEGER DEFAULT NULL,
author_id INTEGER NOT NULL DEFAULT 0,
category_id INTEGER DEFAULT NULL,
thread_id INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
url TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
listed DATETIME DEFAULT NULL,
timezone TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
last_index DATETIME DEFAULT NULL
);

CREATE TABLE linker_category (
id INTEGER PRIMARY KEY AUTOINCREMENT,
position INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
sortorder INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE linker_queue (
id INTEGER PRIMARY KEY AUTOINCREMENT,
language_id INTEGER DEFAULT NULL,
submitter_id INTEGER NOT NULL DEFAULT 0,
category_id INTEGER DEFAULT NULL,
title TEXT DEFAULT NULL,
url TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
timezone TEXT DEFAULT NULL
);

CREATE TABLE linker_search (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
linker_id INTEGER NOT NULL DEFAULT 0,
word_count INTEGER NOT NULL DEFAULT 1,
PRIMARY KEY (searchwordlist_idsearchwordlist,linker_id)
);

CREATE TABLE roles (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
can_login INTEGER NOT NULL DEFAULT 0,
is_admin INTEGER NOT NULL DEFAULT 0,
private_labels INTEGER NOT NULL DEFAULT 1,
public_profile_allowed_at DATETIME DEFAULT NULL,
UNIQUE (name)
);

CREATE TABLE user_roles (
iduser_roles INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL,
role_id INTEGER NOT NULL
);

CREATE TABLE grants (
id INTEGER PRIMARY KEY AUTOINCREMENT,
created_at DATETIME DEFAULT NULL,
updated_at DATETIME DEFAULT NULL,
user_id INTEGER DEFAULT NULL,
role_id INTEGER DEFAULT NULL,
section TEXT NOT NULL,
item TEXT DEFAULT NULL,
rule_type TEXT NOT NULL,
item_id INTEGER DEFAULT NULL,
item_rule TEXT DEFAULT NULL,
action TEXT NOT NULL,
extra TEXT DEFAULT NULL,
active INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE preferences (
idpreferences INTEGER PRIMARY KEY AUTOINCREMENT,
language_id INTEGER DEFAULT NULL,
users_idusers INTEGER NOT NULL DEFAULT 0,
emailforumupdates INTEGER DEFAULT 0,
page_size INTEGER NOT NULL DEFAULT 15,
auto_subscribe_replies INTEGER NOT NULL DEFAULT 1,
timezone TEXT DEFAULT NULL,
custom_css TEXT DEFAULT NULL,
daily_digest_hour int DEFAULT NULL,
daily_digest_mark_read INTEGER NOT NULL DEFAULT 0,
last_digest_sent_at DATETIME DEFAULT NULL,
weekly_digest_day INT DEFAULT NULL,
weekly_digest_hour INT DEFAULT NULL,
last_weekly_digest_sent_at DATETIME DEFAULT NULL,
monthly_digest_day INT DEFAULT NULL,
monthly_digest_hour INT DEFAULT NULL,
last_monthly_digest_sent_at DATETIME DEFAULT NULL,
image_safe_dimension TEXT DEFAULT NULL
);

CREATE TABLE searchwordlist (
idsearchwordlist INTEGER PRIMARY KEY AUTOINCREMENT,
word TEXT DEFAULT NULL,
UNIQUE (word)
);

CREATE TABLE searchwordlist_has_linker (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
linker_id INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (searchwordlist_idsearchwordlist,linker_id)
);

CREATE TABLE site_news (
idsiteNews INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INTEGER NOT NULL DEFAULT 0,
language_id INTEGER DEFAULT NULL,
users_idusers INTEGER NOT NULL DEFAULT 0,
news TEXT DEFAULT NULL,
occurred DATETIME DEFAULT NULL,
timezone TEXT DEFAULT NULL,
last_index DATETIME DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL
);

CREATE TABLE site_news_search (
site_news_id INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
word_count INTEGER NOT NULL DEFAULT 1,
PRIMARY KEY (site_news_id,searchwordlist_idsearchwordlist)
);

CREATE TABLE user_language (
iduserlang INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_id INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE users (
idusers INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
public_profile_enabled_at DATETIME DEFAULT NULL,
UNIQUE (username)
);

CREATE TABLE passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_emails (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id int NOT NULL,
email TEXT NOT NULL DEFAULT '',
verified_at DATETIME DEFAULT NULL,
last_verification_code TEXT DEFAULT NULL,
verification_expires_at DATETIME DEFAULT NULL,
notification_priority int NOT NULL DEFAULT 0,
UNIQUE (email,last_verification_code)
);

CREATE TABLE writing (
idwriting INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
forumthread_id INTEGER NOT NULL DEFAULT 0,
language_id INTEGER DEFAULT NULL,
writing_category_id INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
published DATETIME DEFAULT NULL,
timezone TEXT DEFAULT NULL,
writing TEXT DEFAULT NULL,
abstract TEXT DEFAULT NULL,
private INTEGER DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
last_index DATETIME DEFAULT NULL
);

CREATE TABLE writing_category (
idwritingCategory INTEGER PRIMARY KEY AUTOINCREMENT,
writing_category_id INTEGER DEFAULT NULL,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);

CREATE TABLE writing_search (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
writing_id INTEGER NOT NULL DEFAULT 0,
word_count INTEGER NOT NULL DEFAULT 1,
PRIMARY KEY (searchwordlist_idsearchwordlist,writing_id)
);

CREATE TABLE IF NOT EXISTS api_keys (
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

CREATE TABLE IF NOT EXISTS goose_db_version (
id INTEGER PRIMARY KEY AUTOINCREMENT,
version_id bigint NOT NULL,
is_applied boolean NOT NULL,
tstamp timestamp NULL default CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
method TEXT NOT NULL DEFAULT 'internal',
pattern TEXT NOT NULL DEFAULT '',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pending_emails (
id INTEGER PRIMARY KEY AUTOINCREMENT,
to_user_id int DEFAULT NULL,
direct_email INTEGER NOT NULL DEFAULT 0,
body TEXT NOT NULL,
error_count int NOT NULL DEFAULT 0,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
sent_at DATETIME DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS pending_passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id int NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT NOT NULL,
verification_code TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
verified_at DATETIME DEFAULT NULL,
UNIQUE (verification_code)
);

CREATE TABLE IF NOT EXISTS notifications (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
link TEXT,
message TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
read_at DATETIME DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS dead_letters (
id INTEGER PRIMARY KEY AUTOINCREMENT,
message TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
session_id TEXT NOT NULL,
users_idusers int NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (session_id)
);

CREATE TABLE IF NOT EXISTS site_announcements (
id INTEGER PRIMARY KEY AUTOINCREMENT,
site_news_id int NOT NULL,
active INTEGER NOT NULL DEFAULT 1,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS login_attempts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL,
ip_address TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS banned_ips (
id INTEGER PRIMARY KEY AUTOINCREMENT,
ip_net TEXT NOT NULL,
reason TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
expires_at DATETIME DEFAULT NULL,
canceled_at DATETIME DEFAULT NULL,
UNIQUE (ip_net)
);

CREATE TABLE IF NOT EXISTS template_overrides (
name TEXT NOT NULL,
body TEXT NOT NULL,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS audit_log (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
action TEXT NOT NULL,
path TEXT NOT NULL,
details TEXT,
data TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS deactivated_users (
idusers int NOT NULL,
email TEXT,
passwd TEXT,
passwd_algorithm TEXT,
username TEXT,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idusers)
);

CREATE TABLE IF NOT EXISTS deactivated_comments (
idcomments int NOT NULL,
forumthread_id int NOT NULL,
users_idusers int NOT NULL,
language_id int DEFAULT NULL,
written DATETIME,
TEXT TEXT,
timezone TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idcomments)
);

CREATE TABLE IF NOT EXISTS deactivated_writings (
idwriting int NOT NULL,
users_idusers int NOT NULL,
forumthread_id int NOT NULL,
language_id int DEFAULT NULL,
writing_category_id int NOT NULL,
title TEXT,
published DATETIME,
timezone TEXT DEFAULT NULL,
writing TEXT,
abstract TEXT,
private INTEGER DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idwriting)
);

CREATE TABLE IF NOT EXISTS deactivated_blogs (
idblogs int NOT NULL,
forumthread_id int NOT NULL,
users_idusers int NOT NULL,
language_id int DEFAULT NULL,
blog TEXT,
written DATETIME,
timezone TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idblogs)
);

CREATE TABLE IF NOT EXISTS deactivated_imageposts (
idimagepost int NOT NULL,
forumthread_id int NOT NULL,
users_idusers int NOT NULL,
imageboard_idimageboard int DEFAULT NULL,
posted DATETIME,
timezone TEXT DEFAULT NULL,
description TEXT,
thumbnail TEXT,
fullimage TEXT,
file_size int NOT NULL,
approved INTEGER DEFAULT 0,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idimagepost)
);

CREATE TABLE IF NOT EXISTS deactivated_linker (
id int NOT NULL,
language_id int DEFAULT NULL,
author_id int NOT NULL,
category_id int DEFAULT NULL,
thread_id int NOT NULL,
title TEXT,
url TEXT,
description TEXT,
listed DATETIME,
timezone TEXT DEFAULT NULL,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS uploaded_images (
iduploadedimage INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
path TEXT,
width int DEFAULT NULL,
height int DEFAULT NULL,
file_size int NOT NULL,
uploaded DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS thread_images (
idthread_image INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id int NOT NULL,
path TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_user_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_request_queue (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
change_table TEXT NOT NULL,
change_field TEXT NOT NULL,
change_row_id int NOT NULL,
change_value TEXT,
contact_options TEXT,
status TEXT NOT NULL DEFAULT 'pending',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
acted_at DATETIME DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS admin_request_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
request_id int NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS external_links (
id INTEGER PRIMARY KEY AUTOINCREMENT,
url TEXT NOT NULL,
clicks int NOT NULL DEFAULT 0,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
updated_by int DEFAULT NULL,
card_title TEXT,
card_description TEXT,
card_image TEXT,
card_image_cache TEXT,
favicon_cache TEXT,
card_duration TEXT,
card_upload_date TEXT,
card_author TEXT,
UNIQUE (url)
);

CREATE TABLE content_public_labels (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id int NOT NULL,
label TEXT NOT NULL,
UNIQUE (item,item_id,label)
);

CREATE TABLE content_private_labels (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id int NOT NULL,
user_id int NOT NULL,
label TEXT NOT NULL,
invert INTEGER NOT NULL DEFAULT 0,
UNIQUE (item,item_id,user_id,label)
);

CREATE TABLE content_label_status (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id int NOT NULL,
label TEXT NOT NULL,
UNIQUE (item,item_id,label)
);

CREATE TABLE content_read_markers (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id int NOT NULL,
user_id int NOT NULL,
last_comment_id int NOT NULL,
UNIQUE (item,item_id,user_id)
);

CREATE TABLE role_subscription_archetypes (
id INTEGER PRIMARY KEY AUTOINCREMENT,
role_id int NOT NULL,
archetype_name TEXT NOT NULL,
pattern TEXT NOT NULL,
method TEXT NOT NULL DEFAULT 'internal',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scheduler_state (
task_name TEXT NOT NULL PRIMARY KEY,
last_run_at DATETIME DEFAULT NULL,
metadata TEXT DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS image_cache_entries (
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
uploaded_image_id int DEFAULT NULL,
error_message TEXT DEFAULT NULL,
retry_count int NOT NULL DEFAULT 0,
last_attempt_at DATETIME DEFAULT NULL,
next_attempt_at DATETIME DEFAULT NULL,
PRIMARY KEY (id)
);

INSERT INTO goose_db_version (version_id, is_applied) VALUES (92, 1);

INSERT INTO goose_db_version (version_id, is_applied) VALUES (93, 1);

INSERT INTO goose_db_version (version_id, is_applied) VALUES (94, 1);

CREATE TABLE user_passkeys (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id INT NOT NULL,
name TEXT NOT NULL,
backup_eligible BOOLEAN DEFAULT NULL,
backup_state BOOLEAN DEFAULT NULL,
credential_id BLOB NOT NULL,
public_key BLOB NOT NULL,
attestation_type TEXT NOT NULL,
aaguid BLOB NOT NULL,
sign_count INT NOT NULL DEFAULT 0,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
last_used_at DATETIME DEFAULT NULL,
expires_at DATETIME DEFAULT NULL,
UNIQUE (credential_id)
);

INSERT INTO goose_db_version (version_id, is_applied) VALUES (94, 1);
INSERT INTO goose_db_version (version_id, is_applied) VALUES (95, 1);
