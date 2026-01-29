package entity

import (
	"github.com/fixkme/othello/server/pb/datas"
)

type Player struct {
	*datas.PBPlayerInfo
	GateId       string
	IsRobot      bool
	OfflineTimer int64
}

func NewPlayer(pinfo *datas.PBPlayerInfo) *Player {
	p := &Player{
		PBPlayerInfo: pinfo,
	}
	return p
}

func (p *Player) GetId() int64 {
	return p.Id
}

func (p *Player) GetPlayPieceType() PieceType {
	return PieceType(p.GetPlayPieceType())
}

func (p *Player) SetPlayPieceType(t int64) {
	p.PlayPieceType = t
}
