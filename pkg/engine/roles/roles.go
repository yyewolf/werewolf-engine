package roles

import "github.com/yyewolf/werewolf-engine/pkg/engine"

type Villager struct{}
type Werewolf struct{}
type Seer struct{}
type Witch struct{}
type Hunter struct{}
type LittleGirl struct{}
type Cupid struct{}
type Thief struct{}
type Ancien struct{}
type Scapegoat struct{}
type VillageIdiot struct{}
type FlutePlayer struct{}
type Savior struct{}
type Raven struct{}
type WhiteWolf struct{}
type Pyromaniac struct{}

func (Villager) ID() engine.RoleID     { return engine.RoleVillager }
func (Villager) Team() engine.Team     { return engine.TeamVillagers }
func (Werewolf) ID() engine.RoleID     { return engine.RoleWerewolf }
func (Werewolf) Team() engine.Team     { return engine.TeamWerewolves }
func (Seer) ID() engine.RoleID         { return engine.RoleSeer }
func (Seer) Team() engine.Team         { return engine.TeamVillagers }
func (Witch) ID() engine.RoleID        { return engine.RoleWitch }
func (Witch) Team() engine.Team        { return engine.TeamVillagers }
func (Hunter) ID() engine.RoleID       { return engine.RoleHunter }
func (Hunter) Team() engine.Team       { return engine.TeamVillagers }
func (LittleGirl) ID() engine.RoleID   { return engine.RoleLittleGirl }
func (LittleGirl) Team() engine.Team   { return engine.TeamVillagers }
func (Cupid) ID() engine.RoleID        { return engine.RoleCupid }
func (Cupid) Team() engine.Team        { return engine.TeamVillagers }
func (Thief) ID() engine.RoleID        { return engine.RoleThief }
func (Thief) Team() engine.Team        { return engine.TeamVillagers }
func (Ancien) ID() engine.RoleID       { return engine.RoleAncien }
func (Ancien) Team() engine.Team       { return engine.TeamVillagers }
func (Scapegoat) ID() engine.RoleID    { return engine.RoleScapegoat }
func (Scapegoat) Team() engine.Team    { return engine.TeamVillagers }
func (VillageIdiot) ID() engine.RoleID { return engine.RoleVillageIdiot }
func (VillageIdiot) Team() engine.Team { return engine.TeamVillagers }
func (FlutePlayer) ID() engine.RoleID  { return engine.RoleFlutePlayer }
func (FlutePlayer) Team() engine.Team  { return engine.TeamVillagers }
func (Savior) ID() engine.RoleID       { return engine.RoleSavior }
func (Savior) Team() engine.Team       { return engine.TeamVillagers }
func (Raven) ID() engine.RoleID        { return engine.RoleRaven }
func (Raven) Team() engine.Team        { return engine.TeamVillagers }
func (WhiteWolf) ID() engine.RoleID    { return engine.RoleWhiteWolf }
func (WhiteWolf) Team() engine.Team    { return engine.TeamWerewolves }
func (Pyromaniac) ID() engine.RoleID   { return engine.RolePyromaniac }
func (Pyromaniac) Team() engine.Team   { return engine.TeamVillagers }
