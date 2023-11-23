package interactions

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/bwmarrin/discordgo"
)

type InventoryData struct {
	CharName   string            `json:"charName"`
	ItemCounts map[string]int    `json:"itemCounts"`
	ItemNames  map[string]string `json:"itemNames"`
}

func ParseInventoryData(input string) (*InventoryData, error) {
	b, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}

	inBuf := bytes.NewBuffer(b)

	r, err := zlib.NewReader(inBuf)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	outBuf := bytes.NewBuffer(nil)
	_, err = io.Copy(outBuf, r)
	if err != nil {
		return nil, err
	}

	var result InventoryData
	err = json.Unmarshal(outBuf.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

var loadInventoryCommand = &discordgo.ApplicationCommand{
	Name:        "load-inventory",
	Description: "load-inventory",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "inventory-data",
			Description: "encoded inventory data from GringottsExporter addon",
			Required:    true,
		},
	},
}
