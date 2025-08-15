package internal

import (
	"github.com/fixkme/othello/server/pb/models"
)

type Player struct {
	*models.MPlayerModel
	GateId       string
	PlayingTable *Table
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
