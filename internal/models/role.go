package models

type Role string

const (
	RoleLeader   Role = "leader"
	RoleFollower Role = "follower"
)

func (r Role) Opposite() Role {
	if r == RoleLeader {
		return RoleFollower
	}
	return RoleLeader
}

var Roles = []Role{RoleLeader, RoleFollower}
