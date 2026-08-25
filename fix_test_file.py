import sys

new_test = """
	t.Run("Append Successfully", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: 1, Handler: "forum", Title: sql.NullString{String: "Title", Valid: true},
			}, nil
		}
		q.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
			return &db.GetThreadLastPosterAndPermsForUserRow{
				Idforumthread: 2, ForumtopicIdforumtopic: 1,
			}, nil
		}
		q.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
			return []*db.GetCommentsByThreadIdForUserRow{
				{Idcomments: 10, ForumthreadID: 2, UsersIdusers: 5, Written: sql.NullTime{Time: time.Now().Add(-5 * time.Minute), Valid: true}},
			}, nil
		}
		q.GetCommentByIdFn = func(ctx context.Context, id int32) (*db.Comment, error) {
			return &db.Comment{Idcomments: 10, ForumthreadID: 2, Text: sql.NullString{String: "old text", Valid: true}}, nil
		}

		cfg := config.NewRuntimeConfig()
		cfg.ForumPostAppendWindow = 60

		req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/2/reply", nil)
		req.Form = url.Values{"replytext": {"new text"}, "language": {"1"}}
		req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})

		sess := &sessions.Session{Values: map[any]any{"UID": int32(5)}}
		cd := common.NewCoreData(context.Background(), q, cfg, common.WithSession(sess))
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		q.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
			return 1, nil // 1 row updated
		}

		q.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			t.Fatalf("should not fallback to create comment")
			return 0, nil
		}

		q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		    return 1, nil
		}

		replyTask.Action(httptest.NewRecorder(), req)

		if len(q.AppendCommentInSectionForCommenterCalls) != 1 {
		    t.Fatalf("expected append comment to be called once")
		}

		got := q.AppendCommentInSectionForCommenterCalls[0]
		if got.CommentID != 10 || got.CommenterID != 5 || got.AppendWindowMins != int64(60) {
		    t.Fatalf("unexpected arguments to append: %+v", got)
		}
	})

	t.Run("Fallback on read marker block", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: 1, Handler: "forum", Title: sql.NullString{String: "Title", Valid: true},
			}, nil
		}
		q.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
			return &db.GetThreadLastPosterAndPermsForUserRow{
				Idforumthread: 2, ForumtopicIdforumtopic: 1,
			}, nil
		}
		q.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
			return []*db.GetCommentsByThreadIdForUserRow{
				{Idcomments: 10, ForumthreadID: 2, UsersIdusers: 5, Written: sql.NullTime{Time: time.Now().Add(-5 * time.Minute), Valid: true}},
			}, nil
		}
		q.GetCommentByIdFn = func(ctx context.Context, id int32) (*db.Comment, error) {
			return &db.Comment{Idcomments: 10, ForumthreadID: 2, Text: sql.NullString{String: "old text", Valid: true}}, nil
		}

		cfg := config.NewRuntimeConfig()
		cfg.ForumPostAppendWindow = 60

		req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/2/reply", nil)
		req.Form = url.Values{"replytext": {"new text"}, "language": {"1"}}
		req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})

		sess := &sessions.Session{Values: map[any]any{"UID": int32(5)}}
		cd := common.NewCoreData(context.Background(), q, cfg, common.WithSession(sess))
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		q.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
			return 0, nil // 0 rows updated (e.g. read marker blocked)
		}

		q.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return 11, nil
		}

		q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		    return 1, nil
		}

		replyTask.Action(httptest.NewRecorder(), req)

		if len(q.AppendCommentInSectionForCommenterCalls) != 1 {
		    t.Fatalf("expected append comment to be called once")
		}
		if len(q.CreateCommentInSectionForCommenterCalls) != 1 {
			t.Fatalf("expected fallback to create comment")
		}
	})
"""

with open("handlers/forum/forum_post_append_test.go", "r") as f:
    content = f.read()

content = content.replace("}\n", "}\n" + new_test)

with open("handlers/forum/forum_post_append_test.go", "w") as f:
    f.write(content)
