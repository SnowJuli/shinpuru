package webserver

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type ListResponse struct {
	N    int         `json:"n"`
	Data interface{} `json:"data"`
}

type User struct {
	*discordgo.User

	AvatarURL string `json:"avatar_url"`
}

type Guild struct {
	*discordgo.Guild
	IconURL string `json:"icon_url"`
}

type GuildReduced struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Icon        string              `json:"icon"`
	IconURL     string              `json:"icon_url"`
	Region      string              `json:"region"`
	OwnerID     string              `json:"owner_id"`
	JoinedAt    discordgo.Timestamp `json:"joined_at"`
	MemberCount int                 `json:"member_count"`
}

func GuildFromGuild(g *discordgo.Guild) *Guild {
	return &Guild{
		Guild:   g,
		IconURL: getIconURL(g.ID, g.Icon),
	}
}

func GuildReducedFromGuild(g *discordgo.Guild) *GuildReduced {
	return &GuildReduced{
		ID:          g.ID,
		Name:        g.Name,
		Icon:        g.Icon,
		IconURL:     getIconURL(g.ID, g.Icon),
		Region:      g.Region,
		OwnerID:     g.OwnerID,
		JoinedAt:    g.JoinedAt,
		MemberCount: g.MemberCount,
	}
}

func getIconURL(guildID, iconHash string) string {
	if iconHash == "" {
		return ""
	}

	return fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.png", guildID, iconHash)
}
