package entity

import (
	"github.com/fixkme/gokit/util/time"
	"github.com/fixkme/othello/server/pb/datas"
)

const (
	rowSize         = 8
	colSize         = 8
	pieceType_None  = 0
	pieceType_Black = int8(-1)
	pieceType_White = int8(1)
)

var dirs = [8][2]int8{
	{1, 0}, {-1, 0}, //左右
	{0, -1}, {0, 1}, //上下
	{-1, -1}, {1, 1}, //左斜
	{1, -1}, {-1, 1}, //右斜
}

type Table struct {
	*datas.PBTableInfo
	chesses [rowSize][colSize]int8
}

func (t *Table) Init() {
	if t == nil {
		return
	}
	t.CreatedTime = time.NowMs()
}
