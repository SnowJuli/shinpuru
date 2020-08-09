package middleware

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/internal/core/config"
	"github.com/zekroTJA/shinpuru/internal/core/database"
	"github.com/zekroTJA/shinpuru/internal/core/permissions"
	"github.com/zekroTJA/shinpuru/internal/util/static"
	"github.com/zekroTJA/shinpuru/pkg/discordutil"
	"github.com/zekroTJA/shireikan"
)

type PermissionsMiddleware struct {
	db  database.Database
	cfg *config.Config
}

func NewPermissionMiddleware(db database.Database, cfg *config.Config) *PermissionsMiddleware {
	return &PermissionsMiddleware{db, cfg}
}

func (m *PermissionsMiddleware) Handle(cmd shireikan.Command, ctx shireikan.Context) (next bool, err error) {
	if m.db == nil {
		m.db, _ = ctx.GetObject("db").(database.Database)
	}

	if m.cfg == nil {
		m.cfg, _ = ctx.GetObject("config").(*config.Config)
	}

	var guildID string
	if ctx.GetGuild() != nil {
		guildID = ctx.GetGuild().ID
	}

	ok, _, err := m.CheckPermissions(ctx.GetSession(), guildID, ctx.GetUser().ID, cmd.GetDomainName())

	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return false, err
	}

	if !ok {
		msg, _ := ctx.ReplyEmbedError("You are not permitted to use this command!", "Missing Permission")
		discordutil.DeleteMessageLater(ctx.GetSession(), msg, 8*time.Second)
	}

	return true, nil
}

func (m *PermissionsMiddleware) GetLayer() shireikan.MiddlewareLayer {
	return shireikan.LayerBeforeCommand
}

// GetPermissions tries to fetch the permissions array of
// the passed user of the specified guild. The merged
// permissions array is returned as well as the override,
// which is true when the specified user is the bot owner,
// guild owner or an admin of the guild.
func (m *PermissionsMiddleware) GetPermissions(s *discordgo.Session, guildID, userID string) (perm permissions.PermissionArray, overrideExplicits bool, err error) {
	if guildID != "" {
		perm, err = m.db.GetMemberPermission(s, guildID, userID)
		if err != nil && !database.IsErrDatabaseNotFound(err) {
			return
		}
	} else {
		perm = make(permissions.PermissionArray, 0)
	}

	if m.cfg.Discord.OwnerID == userID {
		perm = perm.Merge(permissions.PermissionArray{"+sp.*"}, false)
		overrideExplicits = true
	}

	if guildID != "" {
		guild, err := discordutil.GetGuild(s, guildID)
		if err != nil {
			return permissions.PermissionArray{}, false, nil
		}

		member, _ := s.GuildMember(guildID, userID)

		if userID == guild.OwnerID || (member != nil && discordutil.IsAdmin(guild, member)) {
			defAdminRoles := m.cfg.Permissions.DefaultAdminRules
			if defAdminRoles == nil {
				defAdminRoles = static.DefaultAdminRules
			}

			perm = perm.Merge(defAdminRoles, false)
			overrideExplicits = true
		}
	}

	defUserRoles := m.cfg.Permissions.DefaultUserRules
	if defUserRoles == nil {
		defUserRoles = static.DefaultUserRules
	}

	perm = perm.Merge(defUserRoles, false)

	return perm, overrideExplicits, nil
}

// CheckPermissions tries to fetch the permissions of the specified user
// on the specified guild and returns true, if the passed dn matches the
// fetched permissions array. Also, the override status is returned as
// well as errors occured during permissions fetching.
func (m *PermissionsMiddleware) CheckPermissions(s *discordgo.Session, guildID, userID, dn string) (bool, bool, error) {
	perms, overrideExplicits, err := m.GetPermissions(s, guildID, userID)
	if err != nil {
		return false, false, err
	}

	return perms.Check(dn), overrideExplicits, nil
}
