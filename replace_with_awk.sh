#!/bin/bash
awk '
/type QuerierStub struct {/ {
    in_struct = 1
    print $0
    next
}
/^}$/ {
    if (in_struct == 1) {
        print "\tListUnreadPrivateThreadsForUserCalls   []ListUnreadPrivateThreadsForUserParams"
        print "\tListUnreadPrivateThreadsForUserReturns []*ListUnreadPrivateThreadsForUserRow"
        print "\tListUnreadPrivateThreadsForUserErr     error"
        print "\tListUnreadPrivateThreadsForUserFn      func(context.Context, ListUnreadPrivateThreadsForUserParams) ([]*ListUnreadPrivateThreadsForUserRow, error)"
        in_struct = 0
    }
    print $0
    next
}
{ print $0 }
END {
    print ""
    print "func (s *QuerierStub) ListUnreadPrivateThreadsForUser(ctx context.Context, arg ListUnreadPrivateThreadsForUserParams) ([]*ListUnreadPrivateThreadsForUserRow, error) {"
    print "\ts.ListUnreadPrivateThreadsForUserCalls = append(s.ListUnreadPrivateThreadsForUserCalls, arg)"
    print "\tfn := s.ListUnreadPrivateThreadsForUserFn"
    print "\tret := s.ListUnreadPrivateThreadsForUserReturns"
    print "\terr := s.ListUnreadPrivateThreadsForUserErr"
    print "\tif fn != nil {"
    print "\t\treturn fn(ctx, arg)"
    print "\t}"
    print "\treturn ret, err"
    print "}"
}
' internal/db/querier_stub.go > temp.go && mv temp.go internal/db/querier_stub.go
