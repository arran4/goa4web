import sys

# We need to add HasOtherUserReadItemAtOrBeyond to sqlite_adapter_gen.go and querier_stub.go
new_adapter = """
func (s *sqliteQuerier) HasOtherUserReadItemAtOrBeyond(ctx context.Context, arg HasOtherUserReadItemAtOrBeyondParams) (bool, error) {
	return s.q.HasOtherUserReadItemAtOrBeyond(ctx, dbsqlite.HasOtherUserReadItemAtOrBeyondParams{
		Item:          arg.Item,
		ItemID:        int64(arg.ItemID),
		UserID:        int64(arg.UserID),
		LastCommentID: int64(arg.LastCommentID),
	})
}
"""

with open("internal/db/sqlite_adapter_gen.go", "a") as f:
    f.write(new_adapter)

new_stub = """
	HasOtherUserReadItemAtOrBeyondFn func(ctx context.Context, arg HasOtherUserReadItemAtOrBeyondParams) (bool, error)
"""
new_stub_method = """
func (s *QuerierStub) HasOtherUserReadItemAtOrBeyond(ctx context.Context, arg HasOtherUserReadItemAtOrBeyondParams) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.HasOtherUserReadItemAtOrBeyondFn == nil {
		panic("QuerierStub.HasOtherUserReadItemAtOrBeyondFn not implemented")
	}
	return s.HasOtherUserReadItemAtOrBeyondFn(ctx, arg)
}
"""

with open("internal/db/querier_stub.go", "r") as f:
    content = f.read()

if "HasOtherUserReadItemAtOrBeyondFn" not in content:
    content = content.replace("type QuerierStub struct {", "type QuerierStub struct {" + new_stub)
    content += new_stub_method
    with open("internal/db/querier_stub.go", "w") as f:
        f.write(content)
