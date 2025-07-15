package entity

import (
	"github.com/fixkme/othello/server/pb/models"
)

type Player struct {
	*models.MPlayerModel
	gateId       string
	PlayingTable *Table
}
