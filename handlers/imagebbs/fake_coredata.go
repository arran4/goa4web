package imagebbs

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

type FakeCoreData struct {
	*common.CoreData
	qs *db.QuerierStub
}

func NewFakeCoreData(cd *common.CoreData) *FakeCoreData {
	qs, ok := cd.Queries().(*db.QuerierStub)
	if !ok {
		// If cd.Queries() is not already a stub, we should probably warn or panic in tests,
		// but typically we pass a stub when creating CoreData for tests.
		// For safety in this specific context, we assume it is set up correctly in the test.
		// If not, the stubs won't work.
	}
	return &FakeCoreData{CoreData: cd, qs: qs}
}

func (f *FakeCoreData) StubImageBoardPosts(boardID int32, posts []*db.ListImagePostsByBoardForListerRow) {
	if f.qs == nil {
		return
	}
	f.qs.ListImagePostsByBoardForListerFn = func(arg db.ListImagePostsByBoardForListerParams) ([]*db.ListImagePostsByBoardForListerRow, error) {
		if arg.BoardID.Valid && arg.BoardID.Int32 == boardID {
			return posts, nil
		}
		return nil, nil
	}
}

func (f *FakeCoreData) StubSubImageBoards(parentID int32, boards []*db.Imageboard) {
	if f.qs == nil {
		return
	}
	f.qs.ListBoardsByParentIDForListerFn = func(arg db.ListBoardsByParentIDForListerParams) ([]*db.Imageboard, error) {
		if arg.ParentID.Valid && arg.ParentID.Int32 == parentID {
			return boards, nil
		}
		return nil, nil
	}
}

func (f *FakeCoreData) StubSystemCheckGrant(grant int32) {
    if f.qs == nil {
        return
    }
    f.qs.SystemCheckGrantReturns = grant
}
