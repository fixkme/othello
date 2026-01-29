package entity

import (
	"github.com/fixkme/gokit/db/mongo/delta"
	"github.com/fixkme/othello/server/pb/models"
)

type Player struct {
	*models.MPlayerModel
	*delta.DeltaCollector[int64] `json:"-"`
	GateId                       string
}

func NewPlayer(id int64, pbModel *models.PBPlayerModel) *Player {
	p := &Player{
		MPlayerModel:   models.NewMPlayerModel(),
		DeltaCollector: delta.NewDeltaCollector(id),
	}
	if pbModel != nil {
		p.InitFromPB(pbModel)
	}
	return p
}

func (p *Player) Id() int64 {
	return p.GetPlayerId()
}

func (p *Player) SetGateId(gateId string) {
	p.GateId = gateId
}
