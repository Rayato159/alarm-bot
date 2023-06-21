package controllers

import (
	"fmt"

	"github.com/Rayato159/awaken-discord-bot/app/models"
	"github.com/Rayato159/awaken-discord-bot/app/repositories"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MemeController struct {
	MemeRepository *repositories.MemeRepository
}

func (h *MemeController) Awaken() {}

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
	}

	if _, err := h.MemeRepository.SetMeme(memeObjectID); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
	}

	meme, err := h.MemeRepository.FindOneMeme(memeObjectID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
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
