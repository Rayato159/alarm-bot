package main

import (
	"os"

	"github.com/Rayato159/awaken-discord-bot/app/server"
	"github.com/Rayato159/awaken-discord-bot/config"
	"github.com/Rayato159/awaken-discord-bot/pkg/database"
)

func main() {
	cfg := config.NewConfig(func() string {
		if len(os.Args) > 1 {
			return os.Args[1]
		} else {
			return "/bin/.env"
		}
	}())

	client := database.DBConnect(cfg)
	defer database.MongoDbDisconnect(client)

	server.NewDiscordServer(cfg, client).Start()
}
