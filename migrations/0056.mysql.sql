ALTER TABLE preferences ADD COLUMN timezone TINYTEXT DEFAULT NULL;

UPDATE blogs
  SET written = COALESCE(
    CONVERT_TZ(written, 'Australia/Melbourne', 'UTC'),
    written
  );

UPDATE comments
  SET written = COALESCE(
    CONVERT_TZ(written, 'Australia/Melbourne', 'UTC'),
    written
  );

UPDATE forumthread
  SET lastaddition = COALESCE(
    CONVERT_TZ(lastaddition, 'Australia/Melbourne', 'UTC'),
    lastaddition
  );

UPDATE forumtopic
  SET lastaddition = COALESCE(
    CONVERT_TZ(lastaddition, 'Australia/Melbourne', 'UTC'),
    lastaddition
  );

UPDATE imagepost
  SET posted = COALESCE(
    CONVERT_TZ(posted, 'Australia/Melbourne', 'UTC'),
    posted
  );

UPDATE linker
  SET listed = COALESCE(
    CONVERT_TZ(listed, 'Australia/Melbourne', 'UTC'),
    listed
  );

UPDATE writing
  SET published = COALESCE(
    CONVERT_TZ(published, 'Australia/Melbourne', 'UTC'),
    published
  );

UPDATE site_news
  SET occurred = COALESCE(
    CONVERT_TZ(occurred, 'Australia/Melbourne', 'UTC'),
    occurred
  );

UPDATE schema_version SET version = 56 WHERE version = 55;
