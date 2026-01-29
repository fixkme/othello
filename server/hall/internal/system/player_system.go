package system

import (
	"context"
	"errors"
	"fmt"

	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/hall/internal/conf"
	"github.com/fixkme/othello/server/hall/internal/entity"
	"github.com/fixkme/othello/server/pb/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type playerSystem struct {
	defaultModule
}

var Player = new(playerSystem)

func init() {
	Manager.register(Player)
}

func (s *playerSystem) onInit() {
}

func (s *playerSystem) afterInit() {
}

func (s *playerSystem) onClose() {
}

func (s *playerSystem) InitModels(p *entity.Player) {
	pinfo := p.GetModelPlayerInfo()
	if pinfo.GetName() == "" {
		pinfo.SetName(fmt.Sprintf("player_%d", p.Id()))
	}
}

func (s *playerSystem) loadData(parentCtx context.Context, playerId int64) (result *models.PBPlayerModel, notFound bool, err error) {
	ctx, cancel := context.WithTimeout(parentCtx, conf.MongoTimeout)
	defer cancel()

	client := core.Mongo.ReadClient()
	coll := client.Database(conf.DBName).Collection(conf.CollPlayer)
	filter := bson.M{"_id": playerId}
	result = models.NewPBPlayerModel()
	if err = coll.FindOne(ctx, filter).Decode(result); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return
		}
		err = nil
		notFound = true
	}

	return
}

func (s *playerSystem) saveData(playerId int64, data any) bool {
	ctx, cancel := context.WithTimeout(context.Background(), conf.MongoTimeout)
	defer cancel()

	client := core.Mongo.WriteClient()
	coll := client.Database(conf.DBName).Collection(conf.CollPlayer)
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": playerId}
	if _, err := coll.UpdateOne(ctx, filter, data, opts); err != nil {
		mlog.Errorf("%d [Player] save data failed, %v; %v", playerId, err, data)
		return false
	}
	return true
}
