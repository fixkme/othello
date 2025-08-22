package internal

import (
	"strconv"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/fixkme/othello/server/pb/models"
)

type Global struct {
	playerIdGen int64
	tableIdGen  int64

	timerCallbacks map[string]func(data any, now int64)

	accPlayers     map[string]int64
	players        map[int64]*Player  // playerId -> Player
	tables         map[int64]*Table   // tableId -> Table
	matching       *redblacktree.Tree // 等待匹配
	offlinePlayers map[int64]int64    // pid => time
}

var global *Global

func init() {
	global = &Global{
		playerIdGen:    0,
		tableIdGen:     0,
		accPlayers:     make(map[string]int64),
		players:        make(map[int64]*Player),
		tables:         make(map[int64]*Table),
		matching:       redblacktree.NewWith(utils.Int64Comparator),
		offlinePlayers: make(map[int64]int64),
	}
}

func (g *Global) GetPlayerByAccount(acc string) *Player {
	id, ok := g.accPlayers[acc]
	if !ok {
		return nil
	}
	p := g.GetPlayer(id)
	return p
}

func (g *Global) GetPlayer(playerId int64) *Player {
	return g.players[playerId]
}

func (g *Global) GetTable(tableId int64) *Table {
	return g.tables[tableId]
}

func (g *Global) CreatePlayer(acc string) *Player {
	pid, ex := g.accPlayers[acc]
	if !ex {
		g.playerIdGen++
		pid = g.playerIdGen
	}

	player := &Player{
		MPlayerModel: models.NewMPlayerModel(),
	}
	player.SetPlayerId(pid)
	pinfo := player.GetModelPlayerInfo()
	pinfo.SetId(pid)
	pinfo.SetAccount(acc)
	pinfo.SetName("player_" + strconv.Itoa(int(pid)))
	g.players[pid] = player
	g.accPlayers[acc] = pid
	return player
}

func (g *Global) RemovePlayer(playerId int64) {
	_, ok := g.players[playerId]
	if !ok {
		return
	}
	delete(g.players, playerId)
}

func (g *Global) CreateTable(p *Player) *Table {
	g.tableIdGen++
	tb := NewTable(g.tableIdGen, p)
	tb.Init()
	g.tables[tb.Id] = tb
	return tb
}

func (g *Global) RemoveTable(tid int64) {
	_, ok := g.tables[tid]
	if !ok {
		return
	}
	delete(g.tables, tid)
}

func (g *Global) CheckMatchTable(p *Player) *Table {
	node := g.matching.Left()
	if node == nil {
		return nil
	}
	tb := node.Value.(*Table)
	g.RemoveMatching(tb.Id)
	tb.MatchPlayer(p)
	msg := &game.PPlayerEnterGame{
		PlayerInfo: p.GetModelPlayerInfo().ToPB(),
	}
	NoticePlayer(msg, tb.OwnerPlayer)
	return tb
}

func (g *Global) AddMatching(tb *Table) {
	g.matching.Put(tb.Id, tb)
}

func (g *Global) RemoveMatching(tid int64) {
	g.matching.Remove(tid)
}

func (g *Global) PlayerLeaveGame(p *Player) error {
	tb := p.PlayingTable
	if tb == nil {
		return nil
	}

	if tb.OppoPlayer == nil {
		// 匹配中离开
		global.RemoveMatching(tb.Id)
		global.RemoveTable(tb.Id)
		p.LeaveGame()
	} else {
		// 游戏中离开、认输
		global.GameOver(tb, p.PlayPieceType(), true)
	}
	return nil
}

func (g *Global) GameOver(tb *Table, loser_piece_type PieceType, isGiveUp bool) {
	msg := &game.PGameResult{
		WinnerPieceType: -int64(loser_piece_type),
		LoserPieceType:  int64(loser_piece_type),
		IsGiveUp:        isGiveUp,
	}
	NoticePlayer(msg, tb.OwnerPlayer, tb.OppoPlayer)

	// 删除
	tb.OwnerPlayer.LeaveGame()
	tb.OppoPlayer.LeaveGame()
	g.RemoveTable(tb.Id)
}
