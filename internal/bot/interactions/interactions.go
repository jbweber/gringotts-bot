package interactions

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jbweber/gringotts-bot/internal/database"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "gbank",
		Description: "Guild Bank",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "search",
				Description: "search",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "name of the item to search for",
						Required:    true,
					},
				},
			},
			{
				Name:        "sniff",
				Description: "sniff",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "name of the item to search for",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "location",
						Description: "name of the item to search for",
						Required:    true,
					},
				},
			},
		},
	},
	{
		Name:        "find-item",
		Description: "find an item in the bank",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "item-name",
				Description: "name of the item to search for",
				Required:    true,
			},
		},
	},
	loadInventoryCommand,
}

//var Handler = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
//	data := i.ApplicationCommandData()
//	switch data.Name {
//	case "hello-world":
//		err := s.InteractionRespond(
//			i.Interaction,
//			&discordgo.InteractionResponse{
//				Type: discordgo.InteractionResponseChannelMessageWithSource,
//				Data: &discordgo.InteractionResponseData{
//					Content: "Hello world!",
//				},
//			},
//		)
//		if err != nil {
//			log.Printf("error occurred, %v", err)
//		}
//		break
//	case "load-inventory":
//		break
//	}
//}

type Handler struct {
	gringotts *database.Gringotts
}

func NewHandler(g *database.Gringotts) *Handler {
	return &Handler{gringotts: g}
}

func (h *Handler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	switch data.Name {
	case "gbank":
		options := i.ApplicationCommandData().Options
		switch options[0].Name {
		case "search":
			h.FindItem2(s, i)
			break
		}
	case "find-item":
		h.FindItem(s, i)
		break
	case "load-inventory":
		h.LoadInventory(s, i)
		break
	default:
		doFailedInteraction(s, i, fmt.Sprintf("unknown command %s", data.Name))
	}
}

func (h *Handler) FindItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.Interaction.ApplicationCommandData().Options

	//if len(opts) > 1 {
	//
	//}

	itemName := opts[0]

	itemNameStr := itemName.Value.(string)

	items, err := h.gringotts.FindItem(context.Background(), itemNameStr)
	if err != nil {
		doFailedInteraction(s, i, fmt.Sprintf("unable to find item: %v", err))
		return
	}

	content := strings.Builder{}
	for _, i := range items {
		content.WriteString(fmt.Sprintf("found %d of item %s with id %s\n", i.Count, getWowheadURL(i.Name, i.ID), i.ID))
	}

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content.String(),
				Flags:   discordgo.MessageFlagsSuppressEmbeds,
			},
		},
	)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}
}

func (h *Handler) FindItem2(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.Interaction.ApplicationCommandData().Options

	//if len(opts) > 1 {
	//
	//}

	itemName := opts[0].Options[0]

	itemNameStr := itemName.Value.(string)

	items, err := h.gringotts.FindItem(context.Background(), itemNameStr)
	if err != nil {
		doFailedInteraction(s, i, fmt.Sprintf("unable to find item: %v", err))
		return
	}

	content := strings.Builder{}
	for _, i := range items {
		content.WriteString(fmt.Sprintf("found %d of item %s with id %s\n", i.Count, getWowheadURL(i.Name, i.ID), i.ID))
	}

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content.String(),
				Flags:   discordgo.MessageFlagsSuppressEmbeds,
			},
		},
	)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}
}

func getWowheadURL(name, id string) string {
	return fmt.Sprintf("[%s](https://www.wowhead.com/classic/item=%s)", name, id)
}

func (h *Handler) LoadInventory(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.Interaction.ApplicationCommandData().Options

	//if len(opts) > 1 {
	//
	//}

	charData := opts[0]

	charDataStr := charData.Value.(string)
	r, err := ParseInventoryData(charDataStr)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}

	err = h.gringotts.UpdateItemCounts(context.Background(), r.CharName, r.ItemCounts)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}

	err = h.gringotts.UpdateItems(context.Background(), r.ItemNames)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("loaded inventory data for %s", r.CharName),
			},
		},
	)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}
}

func doFailedInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
			},
		},
	)
	if err != nil {
		log.Printf("error occurred, %v", err)
	}
}
