package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Rayato159/awaken-discord-bot/app/models"
	"github.com/Rayato159/awaken-discord-bot/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MemeRepository struct {
	Cfg config.IConfig
	Db  *mongo.Client
}

func (r *MemeRepository) InsertMeme(req *models.Meme) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var mutex sync.Mutex
	mutex.Lock()

	result, err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme").InsertOne(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("insert meme failed: %v", err)
	}
	fmt.Printf("The meme has been inserted id: %v\n", result.InsertedID)

	mutex.Unlock()
	return result.InsertedID, nil
}

func (r *MemeRepository) FindOneMeme(memeId any) (*models.Meme, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	result := new(models.Meme)
	if err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme").FindOne(ctx, bson.M{"_id": memeId}, nil).Decode(result); err != nil {
		return nil, fmt.Errorf("find one meme_id: %v failed: %v", memeId, err)
	}
	return result, nil
}

func (r *MemeRepository) FindMeme() ([]*models.Meme, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	cursor, err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme").Find(ctx, bson.D{}, nil)
	if err != nil {
		return nil, fmt.Errorf("find meme failed: %v", err)
	}

	results := make([]*models.Meme, 0)
	for cursor.Next(ctx) {
		result := new(models.Meme)
		err := cursor.Decode(result)
		if err != nil {
			return nil, fmt.Errorf("decode meme failed: %v", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *MemeRepository) SetMeme(memeId primitive.ObjectID) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	type memeToHold struct {
		MemeId primitive.ObjectID `bson:"meme_id,omitempty" json:"meme_id"`
	}

	deletedResult, err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme_to_hold").DeleteMany(ctx, bson.M{}, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(deletedResult)

	result, err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme_to_hold").InsertOne(
		ctx,
		&memeToHold{MemeId: memeId},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("set meme_id: %v failed: %v", memeId, err)
	}
	return result.InsertedID, nil
}

func (r *MemeRepository) Awaken() (*models.Meme, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	pipeline := bson.A{
		bson.M{
			"$limit": 1,
		},
		bson.M{
			"$project": bson.M{
				"_id":     0,
				"meme_id": 1,
			},
		},
	}

	cursor, err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme_to_hold").Aggregate(ctx, pipeline, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type memeToHold struct {
		MemeObjectID primitive.ObjectID `bson:"meme_id,omitempty" json:"meme_id"`
	}

	hold := memeToHold{}
	for cursor.Next(ctx) {
		err := cursor.Decode(&hold)
		if err != nil {
			return nil, fmt.Errorf("decode memeObjectID failed: %v", err)
		}
	}

	result := new(models.Meme)
	if err := r.Db.Database(r.Cfg.Db().Dbname()).Collection("meme").FindOne(ctx, bson.M{"_id": hold.MemeObjectID}, nil).Decode(result); err != nil {
		return nil, fmt.Errorf("find one meme_id: %v failed: %v", hold.MemeObjectID, err)
	}
	return result, nil
}
