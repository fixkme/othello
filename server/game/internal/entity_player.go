package internal

import (
	"github.com/fixkme/gokit/db/mongo/delta"
	"github.com/fixkme/othello/server/pb/models"
)

type Player struct {
	*models.MPlayerModel
	GateId       string
	PlayingTable *Table

	*delta.DeltaCollector[int64] `json:"-"`
}

func NewPlayer(id int64, modelData *models.MPlayerModel) *Player {
	p := &Player{
		DeltaCollector: delta.NewDeltaCollector(id),
	}
	if modelData != nil {
		p.MPlayerModel = modelData
	} else {
		p.MPlayerModel = models.NewMPlayerModel()
	}
	return p
}

func (p *Player) Id() int64 {
	return p.MPlayerModel.GetPlayerId()
}

func (p *Player) PlayPieceType() PieceType {
	return PieceType(p.GetModelPlayerInfo().GetPlayPieceType())
}

func (p *Player) LeaveGame() {
	p.PlayingTable = nil
	p.GetModelPlayerInfo().SetPlayPieceType(0)
}
