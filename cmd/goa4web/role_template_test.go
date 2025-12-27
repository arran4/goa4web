package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

type roleInsertCall struct {
	name           string
	canLogin       bool
	isAdmin        bool
	privateLabels  bool
	returnedRoleID int32
}

type fakeRoleApplier struct {
	roles       map[string]*db.Role
	nextRoleID  int32
	insertCalls []roleInsertCall
	updateCalls []db.AdminUpdateRoleParams
	deleteCalls []sql.NullInt32
	grantCalls  []db.CreateGrantParams
	insertErr   error
}

func (f *fakeRoleApplier) GetRoleByName(_ context.Context, name string) (*db.Role, error) {
	if r, ok := f.roles[name]; ok {
		return r, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeRoleApplier) AdminUpdateRole(_ context.Context, arg db.AdminUpdateRoleParams) error {
	f.updateCalls = append(f.updateCalls, arg)
	if f.roles == nil {
		f.roles = map[string]*db.Role{}
	}
	f.roles[arg.Name] = &db.Role{
		ID:            arg.ID,
		Name:          arg.Name,
		CanLogin:      arg.CanLogin,
		IsAdmin:       arg.IsAdmin,
		PrivateLabels: arg.PrivateLabels,
	}
	return nil
}

func (f *fakeRoleApplier) DeleteGrantsByRoleID(_ context.Context, roleID sql.NullInt32) error {
	f.deleteCalls = append(f.deleteCalls, roleID)
	return nil
}

func (f *fakeRoleApplier) CreateGrant(_ context.Context, arg db.CreateGrantParams) error {
	f.grantCalls = append(f.grantCalls, arg)
	return nil
}

func (f *fakeRoleApplier) RoleInsert(_ context.Context, name string, canLogin, isAdmin, privateLabels bool) (int32, error) {
	if f.nextRoleID == 0 {
		f.nextRoleID = 1
	}
	id := f.nextRoleID
	f.nextRoleID++
	if f.insertErr != nil {
		return 0, f.insertErr
	}
	f.insertCalls = append(f.insertCalls, roleInsertCall{
		name:           name,
		canLogin:       canLogin,
		isAdmin:        isAdmin,
		privateLabels:  privateLabels,
		returnedRoleID: id,
	})
	if f.roles == nil {
		f.roles = map[string]*db.Role{}
	}
	f.roles[name] = &db.Role{ID: id, Name: name, CanLogin: canLogin, IsAdmin: isAdmin, PrivateLabels: privateLabels}
	return id, nil
}

func TestApplyRolesUsesRoleInsertAndUpdates(t *testing.T) {
	ctx := context.Background()
	fake := &fakeRoleApplier{
		roles: map[string]*db.Role{
			"existing": {ID: 5, Name: "existing", CanLogin: true, IsAdmin: false, PrivateLabels: true},
		},
		nextRoleID: 42,
	}

	roles := []RoleDef{
		{
			Name:     "new",
			CanLogin: true,
			IsAdmin:  false,
			Grants: []GrantDef{
				{Section: "news", Item: "post", Action: "see", ItemID: 3},
				{Section: "faq", Item: "", Action: "search"},
			},
		},
		{
			Name:     "existing",
			CanLogin: false,
			IsAdmin:  true,
		},
	}

	if err := applyRoles(ctx, fake, roles); err != nil {
		t.Fatalf("applyRoles: %v", err)
	}

	if len(fake.insertCalls) != 1 {
		t.Fatalf("expected 1 insert call, got %d", len(fake.insertCalls))
	}
	insert := fake.insertCalls[0]
	if insert.name != "new" || !insert.canLogin || insert.isAdmin {
		t.Fatalf("unexpected insert call: %+v", insert)
	}
	if !insert.privateLabels {
		t.Fatalf("expected private labels to follow canLogin for insert, got %#v", insert)
	}

	if len(fake.deleteCalls) != 2 {
		t.Fatalf("expected DeleteGrantsByRoleID called twice, got %d", len(fake.deleteCalls))
	}
	deletes := map[int32]bool{}
	for _, d := range fake.deleteCalls {
		deletes[d.Int32] = d.Valid
	}
	if !deletes[insert.returnedRoleID] {
		t.Fatalf("missing delete for inserted role id %d", insert.returnedRoleID)
	}
	if !deletes[5] {
		t.Fatalf("missing delete for existing role id 5")
	}

	if len(fake.grantCalls) != 2 {
		t.Fatalf("expected 2 grant creations, got %d", len(fake.grantCalls))
	}
	for _, g := range fake.grantCalls {
		if g.RoleID.Int32 != insert.returnedRoleID {
			t.Fatalf("grant created for wrong role id: %#v", g.RoleID)
		}
	}

	if len(fake.updateCalls) != 1 {
		t.Fatalf("expected 1 update call, got %d", len(fake.updateCalls))
	}
	update := fake.updateCalls[0]
	if update.Name != "existing" || update.CanLogin || !update.IsAdmin || update.PrivateLabels {
		t.Fatalf("unexpected update call: %#v", update)
	}
}
