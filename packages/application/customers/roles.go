package customers

import "sort"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Permission string

const (
	PermissionMembersManage      Permission = "members.manage"
	PermissionOrganizationRead   Permission = "organization.read"
	PermissionOrganizationManage Permission = "organization.manage"
	PermissionDevelopersManage   Permission = "developers.manage"
)

var rolePermissions = map[Role]map[Permission]bool{
	RoleOwner:  {PermissionMembersManage: true, PermissionOrganizationRead: true, PermissionOrganizationManage: true, PermissionDevelopersManage: true},
	RoleAdmin:  {PermissionMembersManage: true, PermissionOrganizationRead: true, PermissionDevelopersManage: true},
	RoleMember: {PermissionOrganizationRead: true},
}

func (r Role) Can(p Permission) bool { return rolePermissions[r][p] }

func (r Role) Permissions() []Permission {
	permissions := make([]Permission, 0, len(rolePermissions[r]))
	for permission := range rolePermissions[r] {
		permissions = append(permissions, permission)
	}
	sort.Slice(permissions, func(i, j int) bool { return permissions[i] < permissions[j] })
	return permissions
}
