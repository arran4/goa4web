#!/bin/bash

# Update database/schema.mysql.sql
sed -i '/`locked` tinyint(1) DEFAULT NULL,/a \  `reply_to_comment_id` int(10) DEFAULT NULL,\n  `reply_to_thread_id` int(10) DEFAULT NULL,' database/schema.mysql.sql

# Update test schemas
sed -i '/`locked` tinyint(1) DEFAULT NULL,/a \  `reply_to_comment_id` int(10) DEFAULT NULL,\n  `reply_to_thread_id` int(10) DEFAULT NULL,' internal/app/dbstart/testdata/original.mysql.sql
sed -i '/`locked` tinyint(1) DEFAULT NULL,/a \  `reply_to_comment_id` int(10) DEFAULT NULL,\n  `reply_to_thread_id` int(10) DEFAULT NULL,' testdata/schema/original.mysql.sql

# Update migrations test
sed -i 's/VALUES (89, 1)/VALUES (90, 1)/' migrations/migrations_test.go
