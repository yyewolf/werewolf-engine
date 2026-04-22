package engine

type staticRole struct {
	id   RoleID
	team Team
}

func (r staticRole) ID() RoleID { return r.id }
func (r staticRole) Team() Team { return r.team }

func newRoleByID(roleID RoleID) (Role, error) {
	switch roleID {
	case RoleVillager:
		return staticRole{id: roleID, team: TeamVillagers}, nil
	case RoleWerewolf, RoleWhiteWolf:
		return staticRole{id: roleID, team: TeamWerewolves}, nil
	case RoleSeer, RoleWitch, RoleHunter, RoleLittleGirl, RoleCupid, RoleThief, RoleAncien, RoleScapegoat, RoleVillageIdiot, RoleFlutePlayer, RoleSavior, RoleRaven, RolePyromaniac:
		return staticRole{id: roleID, team: TeamVillagers}, nil
	default:
		return nil, ErrTransitionInvalid
	}
}
