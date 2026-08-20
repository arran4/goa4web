-- +goose Up
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
forumthread_idforumthread INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
blog TEXT DEFAULT NULL,
written DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE blogsSearch (
blogs_idblogs INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (blogs_idblogs,searchwordlist_idsearchwordlist)
);

CREATE TABLE bookmarks (
idbookmarks INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
list BLOB DEFAULT NULL
);

CREATE TABLE comments (
idcomments INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_idforumthread INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
written DATETIME DEFAULT NULL,
TEXT TEXT DEFAULT NULL
);

CREATE TABLE commentsSearch (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
comments_idcomments INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (searchwordlist_idsearchwordlist,comments_idcomments)
);

CREATE TABLE faq (
idfaq INTEGER PRIMARY KEY AUTOINCREMENT,
faqCategories_idfaqCategories INTEGER NOT NULL DEFAULT 0,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
answer TEXT DEFAULT NULL,
question TEXT DEFAULT NULL
);

CREATE TABLE faqCategories (
idfaqCategories INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT DEFAULT NULL
);

CREATE TABLE forumcategory (
idforumcategory INTEGER PRIMARY KEY AUTOINCREMENT,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);

CREATE TABLE forumthread (
idforumthread INTEGER PRIMARY KEY AUTOINCREMENT,
firstpost INTEGER NOT NULL DEFAULT 0,
lastposter INTEGER NOT NULL DEFAULT 0,
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
comments INTEGER DEFAULT NULL,
lastaddition DATETIME DEFAULT NULL,
locked INTEGER DEFAULT NULL
);

CREATE TABLE forumtopic (
idforumtopic INTEGER PRIMARY KEY AUTOINCREMENT,
lastposter INTEGER NOT NULL DEFAULT 0,
forumcategory_idforumcategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
threads INTEGER DEFAULT NULL,
comments INTEGER DEFAULT NULL,
lastaddition DATETIME DEFAULT NULL
);

CREATE TABLE imageboard (
idimageboard INTEGER PRIMARY KEY AUTOINCREMENT,
imageboard_idimageboard INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);

CREATE TABLE imagepost (
idimagepost INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_idforumthread INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
imageboard_idimageboard INTEGER NOT NULL DEFAULT 0,
posted DATETIME DEFAULT NULL,
description TEXT DEFAULT NULL,
thumbnail TEXT DEFAULT NULL,
fullimage TEXT DEFAULT NULL
);

CREATE TABLE imagepostSearch (
imagepost_idimagepost INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (imagepost_idimagepost,searchwordlist_idsearchwordlist)
);

CREATE TABLE language (
idlanguage INTEGER PRIMARY KEY AUTOINCREMENT,
nameof TEXT DEFAULT NULL
);

CREATE TABLE linker (
idlinker INTEGER PRIMARY KEY AUTOINCREMENT,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
linkerCategory_idlinkerCategory INTEGER NOT NULL DEFAULT 0,
forumthread_idforumthread INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
url TEXT DEFAULT NULL,
description TEXT DEFAULT NULL,
listed DATETIME DEFAULT NULL
);

CREATE TABLE linkerCategory (
idlinkerCategory INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT DEFAULT NULL
);

CREATE TABLE linkerQueue (
idlinkerQueue INTEGER PRIMARY KEY AUTOINCREMENT,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
linkerCategory_idlinkerCategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
url TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);

CREATE TABLE linkerSearch (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
linker_idlinker INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (searchwordlist_idsearchwordlist,linker_idlinker)
);

CREATE TABLE permissions (
idpermissions INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
section TEXT DEFAULT NULL,
level tinyblob DEFAULT NULL
);

CREATE TABLE preferences (
idpreferences INTEGER PRIMARY KEY AUTOINCREMENT,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
emailforumupdates INTEGER DEFAULT 0
);

CREATE TABLE searchwordlist (
idsearchwordlist INTEGER PRIMARY KEY AUTOINCREMENT,
word TEXT DEFAULT NULL
);

CREATE TABLE searchwordlist_has_linker (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
linker_idlinker INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (searchwordlist_idsearchwordlist,linker_idlinker)
);

CREATE TABLE siteNews (
idsiteNews INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_idforumthread INTEGER NOT NULL DEFAULT 0,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
news TEXT DEFAULT NULL,
occured DATETIME DEFAULT NULL
);

CREATE TABLE siteNewsSearch (
siteNews_idsiteNews INTEGER NOT NULL DEFAULT 0,
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (siteNews_idsiteNews,searchwordlist_idsearchwordlist)
);

CREATE TABLE topicrestrictions (
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
viewlevel INTEGER DEFAULT NULL,
replylevel INTEGER DEFAULT NULL,
newthreadlevel INTEGER DEFAULT NULL,
seelevel INTEGER DEFAULT NULL,
invitelevel INTEGER DEFAULT NULL,
readlevel INTEGER DEFAULT NULL,
modlevel INTEGER DEFAULT NULL,
adminlevel INTEGER DEFAULT NULL
);

CREATE TABLE userlang (
iduserlang INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
language_idlanguage INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE users (
idusers INTEGER PRIMARY KEY AUTOINCREMENT,
email TEXT DEFAULT NULL,
passwd TEXT DEFAULT NULL,
username TEXT DEFAULT NULL
);

CREATE TABLE userstopiclevel (
users_idusers INTEGER NOT NULL DEFAULT 0,
forumtopic_idforumtopic INTEGER NOT NULL DEFAULT 0,
level INTEGER DEFAULT NULL,
invitemax INTEGER DEFAULT NULL,
PRIMARY KEY (users_idusers,forumtopic_idforumtopic)
);

CREATE TABLE writing (
idwriting INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INTEGER NOT NULL DEFAULT 0,
forumthread_idforumthread INTEGER NOT NULL DEFAULT 0,
language_idlanguage INTEGER NOT NULL DEFAULT 0,
writingCategory_idwritingCategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
published DATETIME DEFAULT NULL,
writting TEXT DEFAULT NULL,
abstract TEXT DEFAULT NULL,
private INTEGER DEFAULT NULL
);

CREATE TABLE writingCategory (
idwritingCategory INTEGER PRIMARY KEY AUTOINCREMENT,
writingCategory_idwritingCategory INTEGER NOT NULL DEFAULT 0,
title TEXT DEFAULT NULL,
description TEXT DEFAULT NULL
);

CREATE TABLE writingSearch (
searchwordlist_idsearchwordlist INTEGER NOT NULL DEFAULT 0,
writing_idwriting INTEGER NOT NULL DEFAULT 0,
PRIMARY KEY (searchwordlist_idsearchwordlist,writing_idwriting)
);

CREATE TABLE writtingApprovedUsers (
writing_idwriting INTEGER NOT NULL DEFAULT 0,
users_idusers INTEGER NOT NULL DEFAULT 0,
readdoc INTEGER DEFAULT NULL,
editdoc INTEGER DEFAULT NULL,
PRIMARY KEY (writing_idwriting,users_idusers)
);

CREATE TABLE IF NOT EXISTS schema_version (
version int NOT NULL
);

INSERT INTO schema_version (version) VALUES (1);

CREATE TABLE IF NOT EXISTS subscriptions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
item_type TEXT NOT NULL,
target_id int NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pending_emails (
id INTEGER PRIMARY KEY AUTOINCREMENT,
to_email TEXT NOT NULL,
subject TEXT NOT NULL,
body TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
sent_at DATETIME DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
link TEXT,
message TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
read_at DATETIME DEFAULT NULL
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
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);