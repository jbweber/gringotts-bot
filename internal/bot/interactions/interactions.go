package interactions

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jbweber/gringotts-bot/internal/database"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "hello-world",
		Description: "basic slash command",
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

	item, err := h.gringotts.FindItem(context.Background(), itemNameStr)
	if err != nil {
		doFailedInteraction(s, i, fmt.Sprintf("unable to find item: %v", err))
		return
	}

	err = s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("found %d of item %s", item.Count, item.Name),
			},
		},
	)
	if err != nil {
		doFailedInteraction(s, i, err.Error())
		return
	}
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
