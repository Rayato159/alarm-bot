package server

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rayato159/awaken-discord-bot/app/controllers"
	"github.com/Rayato159/awaken-discord-bot/app/repositories"
	"github.com/Rayato159/awaken-discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	buffer         = make([][]byte, 0)
)

type IDiscordServer interface {
	Start()
}

type discordServer struct {
	cfg             config.IConfig
	db              *mongo.Client
	dg              *discordgo.Session
	commands        []*discordgo.ApplicationCommand
	commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func NewDiscordServer(cfg config.IConfig, db *mongo.Client) IDiscordServer {
	// Init discord server
	dg, err := discordgo.New("Bot " + cfg.App().GetToken())
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}

	return &discordServer{
		dg:              dg,
		db:              db,
		cfg:             cfg,
		commands:        make([]*discordgo.ApplicationCommand, 0),
		commandHandlers: make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)),
	}
}

func (s *discordServer) GetCommandsHandlers() map[string]func(session *discordgo.Session, i *discordgo.InteractionCreate) {
	return s.commandHandlers
}

func (s *discordServer) Start() {
	s.dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	if err := s.dg.Open(); err != nil {
		log.Fatal("error opening connection,", err)
	}

	handler := &controllers.MemeController{
		MemeRepository: &repositories.MemeRepository{
			Cfg: s.cfg,
			Db:  s.db,
		},
	}

	s.commands = append(
		s.commands,
		&discordgo.ApplicationCommand{
			Name:        "add",
			Description: "To add a new meme",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "title",
					Description: "Your meme in abstract",
					Type:        discordgo.ApplicationCommandOptionString,
				},
				{
					Name:        "image",
					Description: "Your meme image url",
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		},
		&discordgo.ApplicationCommand{
			Name:        "set",
			Description: "To set awaken meme by ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "id",
					Description: "Your meme id with out Object('')",
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		},
		&discordgo.ApplicationCommand{
			Name:        "find",
			Description: "Show all meme list",
			Type:        discordgo.ChatApplicationCommand,
		},
	)

	s.commandHandlers["add"] = handler.InsertMeme
	s.commandHandlers["set"] = handler.SetMeme
	s.commandHandlers["find"] = handler.FindMeme

	// router := mux.NewRouter()
	// router.HandleFunc("/meme", module.AtomikkuModule().Handler().RandMeme).Methods("POST")

	// // Start the HTTP server in a separate Goroutine
	// go func() {
	// 	log.Println("Http server is starting.")
	// 	log.Fatal(http.ListenAndServe(":8080", router))
	// }()

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(s.commands))
	for i, v := range s.commands {
		cmd, err := s.dg.ApplicationCommandCreate(s.dg.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	// Cleanly close down the Discord session.
	defer s.dg.Close()

	// Init handlers
	s.dg.AddHandler(func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := s.GetCommandsHandlers()[i.ApplicationCommandData().Name]; ok {
			h(session, i)
		}
	})

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("🤖 Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.dg.ApplicationCommandDelete(s.dg.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
