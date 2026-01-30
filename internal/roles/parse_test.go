package roles

import "testing"

func TestParseRoleNameFromComment(t *testing.T) {
	sql := []byte(`-- Role: moderator
INSERT INTO roles (name, can_login, is_admin) VALUES ('moderator', 1, 0);`)
	name, err := ParseRoleName(sql)
	if err != nil {
		t.Fatalf("ParseRoleName error: %v", err)
	}
	if name != "moderator" {
		t.Fatalf("expected role name moderator, got %q", name)
	}
}

func TestParseRoleGrants(t *testing.T) {
	sql := []byte(`
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'forum', NULL, 'allow', 'view', 1
FROM roles r
WHERE r.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);`)
	grants, err := ParseRoleGrants(sql)
	if err != nil {
		t.Fatalf("ParseRoleGrants error: %v", err)
	}
	if len(grants) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grants))
	}
	if grants[0].Section != "forum" || grants[0].Action != "view" {
		t.Fatalf("unexpected grant: %#v", grants[0])
	}
	if grants[0].Item.Valid {
		t.Fatalf("expected null item, got %q", grants[0].Item.String)
	}
}
