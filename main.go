package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/jbweber/gringotts-bot/internal/bot/interactions"
	"github.com/jbweber/gringotts-bot/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

const (
	appID    = "APP_ID"
	botToken = "BOT_TOKEN"
	dbPath   = "DB_PATH"
	serverID = "SERVER_ID"
)

func main() {
	theAppID, ok := os.LookupEnv(appID)
	if !ok {
		log.Fatal("unable to lookup appID")
	}

	theBotToken, ok := os.LookupEnv(botToken)
	if !ok {
		log.Fatal("unable to lookup bot token")
	}

	theServerID, ok := os.LookupEnv(serverID)
	if !ok {
		log.Fatal("unable to lookup serverID")
	}

	theDBPath, ok := os.LookupEnv(dbPath)
	if !ok {
		log.Fatalf("unable to lookup dbPath")
	}

	s, _ := discordgo.New("Bot " + theBotToken)

	registeredCommands, err := s.ApplicationCommandBulkOverwrite(theAppID, theServerID, interactions.Commands)
	if err != nil {
		log.Fatalf("error registering commands, %v", err)
	}

	db, err := database.NewDB(theDBPath)
	if err != nil {
		log.Fatalf("error creating Database: %v", err)
	}

	defer func() { _ = db.Close() }()

	migrator := database.NewMigrator(db)
	err = migrator.Migrate()
	if err != nil {
		log.Fatalf("error migrating Database: %v", err)
	}

	g := database.NewGringotts(db)

	h := interactions.NewHandler(g)

	s.AddHandler(h.Handle)

	err = s.Open()
	if err != nil {
		log.Fatal(err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, theServerID, v.ID)
		if err != nil {
			log.Printf("error deleting command %s:%s, %v", v.ID, v.Name, err)
		}
	}

	err = s.Close()
	if err != nil {
		log.Fatal(err)
	}
}
