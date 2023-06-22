package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Rayato159/awaken-discord-bot/app/models"
	"github.com/Rayato159/awaken-discord-bot/app/repositories"
	"github.com/Rayato159/awaken-discord-bot/pkg/kawaiiauth"
	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MemeController struct {
	Session        *discordgo.Session
	MemeRepository *repositories.MemeRepository
}

func (h *MemeController) Awaken(c echo.Context) error {
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	if token == "" {
		// Token is missing, return an error response
		return c.JSON(http.StatusUnauthorized, struct {
			Message string `json:"message"`
		}{
			Message: "ไม่ให้เข้า",
		})
	}

	kawaiiAuth := kawaiiauth.NewKawaiiAuth(h.MemeRepository.Cfg)
	_, err := kawaiiAuth.ParseJwtToken(context.Background(), token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, struct {
			Message string `json:"message"`
		}{
			Message: "ไม่ให้เข้า",
		})
	}

	meme, err := h.MemeRepository.Awaken()
	if err != nil {
		return c.JSON(http.StatusUnauthorized, struct {
			Message string `json:"message"`
		}{
			Message: err.Error(),
		})
	}

	embed := &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: meme.ImageUrl,
		},
		Type: discordgo.EmbedTypeImage,
	}

	file, err := os.Open("./bin/assets/andriod-alarm.mp3")
	if err != nil {
		return err
	}
	defer file.Close()

	// Send the message with the embed
	// 419106310110576642 main ch
	channelIds := []string{
		"419106310110576642",
	}

	jobsCh := make(chan string, len(channelIds))
	resultsCh := make(chan string, len(channelIds))

	for _, c := range channelIds {
		jobsCh <- c
	}
	close(jobsCh)

	numberWorkers := 2
	for w := 0; w < numberWorkers; w++ {
		go func(jobsCh <-chan string, resultsCh chan<- string) {
			for job := range jobsCh {
				for i := 0; i < 1; i++ {
					h.Session.ChannelMessageSendEmbed(job, embed)
					h.Session.ChannelMessageSendTTS(job, meme.Title)
					h.Session.ChannelFileSend(job, "andriod-alarm.mp3", file)
				}
				resultsCh <- job
			}
		}(jobsCh, resultsCh)
	}

	// Send the message with the embed
	for a := 0; a < len(channelIds); a++ {
		result := <-resultsCh
		fmt.Println("complated, message sended embed to channel_id:", result)
	}

	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "OK ผ่าน",
	})
}

func (h *MemeController) InsertMeme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command := i.ApplicationCommandData()

	// UserIds
	approvedIds := map[string]bool{
		"272256561366433792": true,
	}

	if !approvedIds[i.Member.User.ID] {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "จ๊ะเอ๊ตัวเอง ฮ่าๆ สวัสดีครับท่านผู้เจริญ",
			},
		})
		return
	}

	req := new(models.Meme)
	for _, c := range command.Options {
		switch c.Name {
		case "title":
			req.Title = c.StringValue()
		case "image":
			req.ImageUrl = c.StringValue()
		}
	}

	if req.Title == "" || req.ImageUrl == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "insert meme failed: title and image are required.",
			},
		})
		return
	}

	memeId, err := h.MemeRepository.InsertMeme(req)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	meme, err := h.MemeRepository.FindOneMeme(memeId)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("insert meme: %s -> completed", meme.Title),
		},
	})

	embed := &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: meme.ImageUrl,
		},
		Title: meme.Title,
		Type:  discordgo.EmbedTypeImage,
	}

	// Send the message with the embed
	s.ChannelMessageSendEmbed(i.ChannelID, embed)
}

func (h *MemeController) SetMeme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command := i.ApplicationCommandData()

	approvedIds := map[string]bool{
		"272256561366433792": true,
	}

	if !approvedIds[i.Member.User.ID] {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "จ๊ะเอ๊ตัวเอง ฮ่าๆ สวัสดีครับท่านผู้เจริญ",
			},
		})
		return
	}

	var memeId string
	for _, c := range command.Options {
		switch c.Name {
		case "id":
			memeId = c.StringValue()
		}
	}
	memeObjectID, err := primitive.ObjectIDFromHex(memeId)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	if _, err := h.MemeRepository.SetMeme(memeObjectID); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	meme, err := h.MemeRepository.FindOneMeme(memeObjectID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("set meme completed"),
		},
	})

	embed := &discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: meme.ImageUrl,
		},
		Title: meme.Title,
		Type:  discordgo.EmbedTypeImage,
	}

	// Send the message with the embed
	s.ChannelMessageSendEmbed(i.ChannelID, embed)
}

func (h *MemeController) FindMeme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	memes, err := h.MemeRepository.FindMeme()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Here, All meme I've.",
		},
	})

	var memeList string
	memeList += "```\n"
	for i, m := range memes {
		memeList += fmt.Sprintf("id:\t%v\n", m.Id)
		memeList += fmt.Sprintf("title:\t%v\n", m.Title)
		memeList += fmt.Sprintf("image_url:\t%v", m.ImageUrl)

		if i != len(memes)-1 {
			memeList += "\n\n"
		}
	}
	memeList += "\n```"
	s.ChannelMessageSend(i.ChannelID, memeList)
}
