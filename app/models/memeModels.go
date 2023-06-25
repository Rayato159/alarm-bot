package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Meme struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title    string             `bson:"title,omitempty" json:"title"`
	ImageUrl string             `bson:"image_url,omitempty" json:"image_url"`
}

type Channels struct {
	ChannelId string `bson:"channel_id,omitempty" json:"channel_ids"`
}
