package internal

import (
	"errors"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/pb/datas"
	"github.com/fixkme/othello/server/pb/game"
)

type PieceType int8

const (
	rowSize         = 8
	colSize         = 8
	PieceType_None  = PieceType(0)
	PieceType_Black = PieceType(-1)
	PieceType_White = PieceType(1)
)

type Piece struct {
	X    int
	Y    int
	Type PieceType
}

func NewPiece(x, y int, t PieceType) *Piece {
	return &Piece{x, y, t}
}

type Vec2 struct {
	X int
	Y int
}

func NewVec2(x, y int) *Vec2 {
	return &Vec2{x, y}
}

var dirs = [8][2]int{
	{1, 0}, {-1, 0}, //左右
	{0, -1}, {0, 1}, //上下
	{-1, -1}, {1, 1}, //左斜
	{1, -1}, {-1, 1}, //右斜
}

var weights = [][]int{
	{8, 1, 5, 1, 1, 5, 1, 8},
	{1, -5, 1, 1, 1, 1, -5, 1},
	{5, 1, 5, 2, 2, 5, 1, 5},
	{1, 1, 2, 1, 1, 2, 1, 1},
	{1, 1, 2, 1, 1, 2, 1, 1},
	{5, 1, 5, 2, 2, 5, 1, 5},
	{1, -5, 1, 1, 1, 1, -5, 1},
	{8, 1, 5, 1, 1, 5, 1, 8},
}

type Table struct {
	//*datas.PBTableInfo
	Id          int64
	OwnerPlayer *Player
	OppoPlayer  *Player
	Chesses     *[rowSize][colSize]PieceType
	Previous    *[rowSize][colSize]PieceType
	Operator    PieceType
	BlackCount  int32
	WhiteCount  int32
	CreateTime  int64
}

func NewTable(tid int64, p *Player) *Table {
	tb := &Table{
		Id:          tid,
		OwnerPlayer: p,
		Chesses:     &[rowSize][colSize]PieceType{},
		Previous:    nil,
	}
	p.PlayingTable = tb
	p.GetModelPlayerInfo().SetPlayPieceType(int64(PieceType_Black)) //创建房间的人为黑棋
	return tb
}

func (tb *Table) Init() {
	tb.CreateTime = util.NowMs()
	tb.Reset()
}

func (tb *Table) PackPB() *datas.PBTableInfo {
	info := &datas.PBTableInfo{
		Id:          tb.Id,
		Turn:        int32(tb.Operator),
		BlackCount:  tb.BlackCount,
		WhiteCount:  tb.WhiteCount,
		CreatedTime: tb.CreateTime,
		Pieces:      make([]*datas.PBPieceInfo, 0, rowSize*colSize),
	}
	if tb.OwnerPlayer != nil {
		info.OwnerPlayer = tb.OwnerPlayer.GetModelPlayerInfo().ToPB()
	}
	if tb.OppoPlayer != nil {
		info.OppoPlayer = tb.OppoPlayer.GetModelPlayerInfo().ToPB()
	}
	for i := 0; i < rowSize; i++ {
		for j := 0; j < colSize; j++ {
			info.Pieces = append(info.Pieces, &datas.PBPieceInfo{X: int32(i), Y: int32(j), Color: int32(tb.Chesses[i][j])})
		}
	}
	return info
}

func (tb *Table) MatchPlayer(p *Player) {
	p.GetModelPlayerInfo().SetPlayPieceType(int64(PieceType_White))
	p.PlayingTable = tb
	tb.OppoPlayer = p
}

func (tb *Table) PlacePiece(x, y int, t PieceType) error {
	if !tb.IsOperator(t) {
		return errors.New("operator not match")
	}
	if !tb.CanPlacePiece(x, y, t) {
		return errors.New("location cant reverse piece")
	}

	// 放下棋子
	//tb.Record()
	tb.AddPiece(x, y, t)
	tb.Reverse(x, y, t)
	// 对方能否有棋可翻转
	if tb.CanPlace(-t) {
		tb.TurnOperator()
		mlog.Debug("change operator to %d", tb.Operator)
	} else {
		mlog.Debug("%d no piece can reverse", -t)
	}
	// 广播落子结果
	msg := &game.PPlacePiece{
		PieceType: int32(t), X: int32(x), Y: int32(y), OperatePiece: int32(tb.Operator),
	}
	NoticePlayer(msg, tb.OwnerPlayer, tb.OppoPlayer)
	// 检查游戏是否达到结束条件
	if tb.CheckEnd() {
		var loser_piece_type PieceType
		if tb.BlackCount > tb.WhiteCount {
			loser_piece_type = PieceType_White
		} else if tb.BlackCount < tb.WhiteCount {
			loser_piece_type = PieceType_Black
		} else {
			loser_piece_type = PieceType_None
		}
		global.GameOver(tb, loser_piece_type, false)
	}

	return nil
}

func (tb *Table) Reset() {
	for i := 0; i < rowSize; i++ {
		for j := 0; j < colSize; j++ {
			tb.Chesses[i][j] = PieceType_None
		}
	}
	tb.BlackCount = 0
	tb.WhiteCount = 0
	tb.AddPiece(3, 3, PieceType_Black)
	tb.AddPiece(3, 4, PieceType_White)
	tb.AddPiece(4, 3, PieceType_White)
	tb.AddPiece(4, 4, PieceType_Black)
	tb.SetOperator(PieceType_Black)
}

func (tb *Table) AddPiece(i, j int, t PieceType) {
	tb.Chesses[i][j] = t
	tb.AddPieceCount(t, 1)
}

func (tb *Table) RemovePiece(i, j int) {
	t := tb.Chesses[i][j]
	tb.Chesses[i][j] = PieceType_None
	tb.AddPieceCount(t, -1)
}

func (tb *Table) AddPieceCount(t PieceType, add int) {
	switch t {
	case PieceType_Black:
		tb.BlackCount += int32(add)
	case PieceType_White:
		tb.WhiteCount += int32(add)
	}
}

// 放置的位置是否合法
func (tb *Table) CanPlaceLocation(i int, j int) bool {
	return tb.InMap(i, j) && tb.Chesses[i][j] == PieceType_None
}

// 在i，j处放置棋子t， 是否能翻转
func (tb *Table) CanPlacePiece(i, j int, t PieceType) bool {
	if !tb.CanPlaceLocation(i, j) {
		return false
	}
	op := -t
	for d := 0; d < len(dirs); d++ {
		di := dirs[d][0]
		dj := dirs[d][1]
		x := i + di
		y := j + dj
		for tb.InMap(x, y) && tb.Chesses[x][y] != PieceType_None {
			if tb.Chesses[x][y] == op {
				x += di
				y += dj
			} else {
				if x == i+di && y == j+dj {
					break
				} else {
					return true
				}
			}
		}
	}
	return false
}

// 当前棋盘能否放置 t 棋子
func (tb *Table) CanPlace(t PieceType) bool {
	for i := range rowSize {
		for j := range colSize {
			if tb.CanPlacePiece(i, j, t) {
				return true
			}
		}
	}
	return false
}

func (tb *Table) GetCanPlacePieceCount(i int, j int, t PieceType) int {
	if !tb.CanPlaceLocation(i, j) {
		return 0
	}
	total := 0
	op := -t
	for d := 0; d < len(dirs); d++ {
		di := dirs[d][0]
		dj := dirs[d][1]
		x := i + di
		y := j + dj
		getNum := 0
		for tb.InMap(x, y) && tb.Chesses[x][y] != PieceType_None {
			if tb.Chesses[x][y] == op {
				getNum++
				x += di
				y += dj
			} else {
				if !(x == i+di && y == j+dj) {
					total += getNum
				}
				break
			}
		}
	}
	return total
}

func (tb *Table) GetBestLocation(t PieceType, difficulty int) *Vec2 {
	loc, _ := tb.GetMax(t, difficulty)
	if loc != nil {
		return loc
	} else {
		return nil
	}
}

func (tb *Table) GetMax(t PieceType, deep int) (*Vec2, float64) {
	if deep == 0 {
		return nil, 0
	}
	op := -t
	max := float64(-65535)
	var loc *Vec2
	for i := 0; i < rowSize; i++ {
		for j := 0; j < colSize; j++ {
			if !tb.CanPlaceLocation(i, j) {
				continue
			}
			tb.AddPiece(i, j, t)
			list := tb.Reverse(i, j, t)
			num := float64(len(list))
			if num > 0 {
				if loc == nil {
					loc = NewVec2(i, j)
				}
				a := float64(tb.BlackCount+tb.WhiteCount) / 64.
				num = a*num + (1-a)*float64(weights[i][j])
				if tb.CheckEnd() {
					num = 10000
				} else {
					_loc, _max := tb.GetMax(op, deep-1)
					if _loc != nil {
						num = num - float64(_max)
					}
				}
				if num > max {
					max = num
					loc.X = i
					loc.Y = j
				}
			}
			// 还原
			tb.RemovePiece(i, j)
			if len(list) > 0 {
				for k := 0; k < len(list); k++ {
					item := list[k]
					tb.Chesses[item.X][item.Y] = op
				}
				tb.AddPieceCount(t, -len(list))
				tb.AddPieceCount(op, len(list))
			}
		}
	}
	if loc == nil {
		return nil, 0
	}
	return loc, max
}

func (tb *Table) TurnOperator() {
	tb.Operator = -tb.Operator
}

func (tb *Table) SetOperator(t PieceType) {
	tb.Operator = t
}

func (tb *Table) IsOperator(t PieceType) bool {
	return tb.Operator == t
}

func (tb *Table) Reverse(i int, j int, t PieceType) (list []*Piece) {
	op := -t
	for d := 0; d < len(dirs); d++ {
		di := dirs[d][0]
		dj := dirs[d][1]
		x := i + di
		y := j + dj
		for tb.InMap(x, y) && tb.Chesses[x][y] != PieceType_None {
			if tb.Chesses[x][y] == op {
				x += di
				y += dj
			} else {
				x -= di
				y -= dj
				for !(x == i && y == j) {
					tb.Chesses[x][y] = t
					list = append(list, NewPiece(x, y, op))
					x -= di
					y -= dj
				}
				break
			}
		}
	}
	if len(list) == 0 {
		return
	}
	tb.AddPieceCount(t, len(list))
	tb.AddPieceCount(op, -len(list))
	return list
}

func (tb *Table) CheckEnd() bool {
	return tb.BlackCount == 0 || tb.WhiteCount == 0 || tb.BlackCount+tb.WhiteCount == 64
}

func (tb *Table) InMap(i int, j int) bool {
	return 0 <= i && i < rowSize && 0 <= j && j < colSize
}

func (tb *Table) Record() {
	tb.Previous = &[rowSize][colSize]PieceType{}
	for i := 0; i < rowSize; i++ {
		for j := 0; j < colSize; j++ {
			tb.Previous[i][j] = tb.Chesses[i][j]
		}
	}
}

func (tb *Table) Undo() (removes, changes []*Piece) {
	if tb.Previous == nil || len(tb.Previous) == 0 {
		return
	}
	tb.BlackCount = 0
	tb.WhiteCount = 0
	var nowPt, oldPt PieceType
	for i := 0; i < rowSize; i++ {
		for j := 0; j < colSize; j++ {
			nowPt, oldPt = tb.Chesses[i][j], tb.Previous[i][j]
			if nowPt != oldPt {
				if nowPt != PieceType_None && oldPt == PieceType_None {
					removes = append(removes, NewPiece(i, j, nowPt))
				} else {
					changes = append(changes, NewPiece(i, j, oldPt))
				}
			}
			tb.Chesses[i][j] = oldPt
			tb.AddPieceCount(tb.Chesses[i][j], 1)
		}
	}
	tb.Previous = nil
	return
}
