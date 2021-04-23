package commands

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/internal/config"
	"github.com/zekroTJA/shinpuru/internal/models"
	"github.com/zekroTJA/shinpuru/internal/services/database"
	"github.com/zekroTJA/shinpuru/internal/services/storage"
	"github.com/zekroTJA/shinpuru/internal/util"
	"github.com/zekroTJA/shinpuru/internal/util/imgstore"
	"github.com/zekroTJA/shinpuru/internal/util/report"
	"github.com/zekroTJA/shinpuru/internal/util/snowflakenodes"
	"github.com/zekroTJA/shinpuru/internal/util/static"
	"github.com/zekroTJA/shinpuru/pkg/acceptmsg"
	"github.com/zekroTJA/shinpuru/pkg/fetch"
	"github.com/zekroTJA/shinpuru/pkg/roleutil"
	"github.com/zekroTJA/shireikan"
)

type CmdKick struct {
}

func (c *CmdKick) GetInvokes() []string {
	return []string{"kick", "userkick"}
}

func (c *CmdKick) GetDescription() string {
	return "Kick users with creating a report entry."
}

func (c *CmdKick) GetHelp() string {
	return "`kick <UserResolvable> <Reason>`"
}

func (c *CmdKick) GetGroup() string {
	return shireikan.GroupModeration
}

func (c *CmdKick) GetDomainName() string {
	return "sp.guild.mod.kick"
}

func (c *CmdKick) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

func (c *CmdKick) IsExecutableInDMChannels() bool {
	return false
}

func (c *CmdKick) Exec(ctx shireikan.Context) error {
	if len(ctx.GetArgs()) < 2 {
		return util.SendEmbedError(ctx.GetSession(), ctx.GetChannel().ID,
			"Invalid command arguments. Please use `help kick` to see how to use this command.").
			DeleteAfter(8 * time.Second).Error()
	}
	victim, err := fetch.FetchMember(ctx.GetSession(), ctx.GetGuild().ID, ctx.GetArgs().Get(0).AsString())
	if err != nil || victim == nil {
		return util.SendEmbedError(ctx.GetSession(), ctx.GetChannel().ID,
			"Sorry, could not find any member :cry:").
			DeleteAfter(8 * time.Second).Error()
	}

	if victim.User.ID == ctx.GetUser().ID {
		return util.SendEmbedError(ctx.GetSession(), ctx.GetChannel().ID,
			"You can not kick yourself...").
			DeleteAfter(8 * time.Second).Error()
	}

	authorMemb, err := ctx.GetSession().GuildMember(ctx.GetGuild().ID, ctx.GetUser().ID)
	if err != nil {
		return err
	}

	if roleutil.PositionDiff(victim, authorMemb, ctx.GetGuild()) >= 0 {
		return util.SendEmbedError(ctx.GetSession(), ctx.GetChannel().ID,
			"You can only kick members with lower permissions than yours.").
			DeleteAfter(8 * time.Second).Error()
	}

	repMsg := strings.Join(ctx.GetArgs()[1:], " ")
	var repType int
	for i, v := range models.ReportTypes {
		if v == "KICK" {
			repType = i
		}
	}
	repID := snowflakenodes.NodesReport[repType].Generate()

	var attachment string
	repMsg, attachment = imgstore.ExtractFromMessage(repMsg, ctx.GetMessage().Attachments)
	if attachment != "" {
		img, err := imgstore.DownloadFromURL(attachment)
		if err == nil && img != nil {
			st, _ := ctx.GetObject(static.DiObjectStorage).(storage.Storage)
			err = st.PutObject(static.StorageBucketImages, img.ID.String(),
				bytes.NewReader(img.Data), int64(img.Size), img.MimeType)
			if err != nil {
				return err
			}
			attachment = img.ID.String()
		}
	}

	cfg, _ := ctx.GetObject(static.DiConfig).(*config.Config)
	db, _ := ctx.GetObject(static.DiDatabase).(database.Database)

	acceptMsg := acceptmsg.AcceptMessage{
		Embed: &discordgo.MessageEmbed{
			Color:       models.ReportColors[repType],
			Title:       "Kick Check",
			Description: "Is everything okay so far?",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Victim",
					Value: fmt.Sprintf("<@%s> (%s#%s)",
						victim.User.ID, victim.User.Username, victim.User.Discriminator),
				},
				{
					Name:  "ID",
					Value: repID.String(),
				},
				{
					Name:  "Type",
					Value: models.ReportTypes[repType],
				},
				{
					Name:  "Description",
					Value: repMsg,
				},
			},
			Image: &discordgo.MessageEmbedImage{
				URL: imgstore.GetLink(attachment, cfg.WebServer.PublicAddr),
			},
		},
		Session:        ctx.GetSession(),
		UserID:         ctx.GetUser().ID,
		DeleteMsgAfter: true,
		AcceptFunc: func(msg *discordgo.Message) {
			rep, err := report.PushKick(
				ctx.GetSession(),
				db,
				cfg.WebServer.PublicAddr,
				ctx.GetGuild().ID,
				ctx.GetUser().ID,
				victim.User.ID,
				repMsg,
				attachment)

			if err != nil {
				util.SendEmbedError(ctx.GetSession(), ctx.GetChannel().ID,
					"Failed kicking member: ```\n"+err.Error()+"\n```")
			} else {
				ctx.GetSession().ChannelMessageSendEmbed(ctx.GetChannel().ID, rep.AsEmbed(cfg.WebServer.PublicAddr))
			}
		},
	}

	_, err = acceptMsg.Send(ctx.GetChannel().ID)

	return err
}
