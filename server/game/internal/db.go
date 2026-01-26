package internal

import (
	"context"
	"time"

	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/pb/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName         = "othello_db"
	playerCollName = "player"
	idCollName     = "id_gen"
	idName_Player  = "player_id_gen"
	idName_Table   = "table_id_gen"
)

type IdSeq struct {
	Id  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

func (g *Global) loadDatas() error {
	coll := core.Mongo.Client().Database(dbName).Collection(playerCollName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		pbData := &models.PBPlayerModel{}
		err := cur.Decode(pbData)
		if err != nil {
			return err
		}
		md := models.NewMPlayerModel()
		md.InitFromPB(pbData)
		p := NewPlayer(pbData.PlayerId, md)
		g.players[pbData.PlayerId] = p
		g.accPlayers[pbData.Account] = pbData.PlayerId
	}
	mlog.Infof("load player data finished, size:%d", len(g.players))
	return nil
}

func (g *Global) GeneId(idName string) (id int64, err error) {
	coll := core.Mongo.Client().Database(dbName).Collection(idCollName)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	after := options.After
	upsert := true
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	filter := bson.M{"_id": idName}
	result := coll.FindOneAndUpdate(ctx, filter, bson.M{"$inc": bson.M{"seq": 1}}, &opt)
	if err = result.Err(); err != nil {
		return
	}
	idSeq := &IdSeq{}
	err = result.Decode(idSeq)
	if err != nil {
		return
	}
	id = idSeq.Seq
	return
}
