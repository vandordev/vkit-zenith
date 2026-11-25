package customers

import "testing"

func TestRolePermissionsMatrix(t *testing.T) {
	want := map[Role][]Permission{
		RoleOwner:  {PermissionDevelopersManage, PermissionMembersManage, PermissionOrganizationManage, PermissionOrganizationRead},
		RoleAdmin:  {PermissionDevelopersManage, PermissionMembersManage, PermissionOrganizationRead},
		RoleMember: {PermissionOrganizationRead},
	}
	for role, expected := range want {
		got := role.Permissions()
		if len(got) != len(expected) {
			t.Errorf("%s permissions = %v, want %v", role, got, expected)
			continue
		}
		for i := range expected {
			if got[i] != expected[i] {
				t.Errorf("%s permissions = %v, want sorted %v", role, got, expected)
				break
			}
		}
	}
}

func TestRolePermissions(t *testing.T) {
	for _, tc := range []struct {
		role       Role
		permission Permission
		want       bool
	}{
		{RoleOwner, PermissionMembersManage, true}, {RoleAdmin, PermissionMembersManage, true}, {RoleMember, PermissionMembersManage, false},
		{RoleOwner, PermissionOrganizationRead, true}, {RoleAdmin, PermissionOrganizationRead, true}, {RoleMember, PermissionOrganizationRead, true},
		{RoleOwner, PermissionOrganizationManage, true}, {RoleAdmin, PermissionOrganizationManage, false}, {RoleMember, PermissionOrganizationManage, false},
		{Role("unknown"), PermissionOrganizationRead, false}, {RoleOwner, Permission("unknown"), false}, {Role(""), PermissionMembersManage, false},
	} {
		if got := tc.role.Can(tc.permission); got != tc.want {
			t.Errorf("%s.Can(%s) = %t, want %t", tc.role, tc.permission, got, tc.want)
		}
	}
}
