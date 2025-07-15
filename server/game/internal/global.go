package internal

import (
	"strconv"

	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/pb/models"
)

type Global struct {
	playerIdGen int64
	tableIdGen  int64

	players  map[int64]*entity.Player // playerId -> Player
	tables   map[int64]*entity.Table
	matching map[int64]*entity.Table // 等待匹配
}

var global *Global

func init() {
	global = &Global{
		playerIdGen: 0,
		tableIdGen:  0,
		players:     make(map[int64]*entity.Player),
		tables:      make(map[int64]*entity.Table),
		matching:    make(map[int64]*entity.Table),
	}
}

func (g *Global) GetPlayer(playerId int64) *entity.Player {
	return g.players[playerId]
}

func (g *Global) GetTable(tableId int64) *entity.Table {
	return g.tables[tableId]
}

func (g *Global) CreatePlayer() *entity.Player {
	g.playerIdGen++
	pid := g.playerIdGen
	player := &entity.Player{
		MPlayerModel: models.NewMPlayerModel(),
	}
	player.SetPlayerId(pid)
	pinfo := player.GetModelPlayerInfo()
	pinfo.SetId(pid)
	pinfo.SetName("player_" + strconv.Itoa(int(pid)))
	g.players[pid] = player
	return player
}
func (g *Global) RemovePlayer(playerId int64) {
	delete(g.players, playerId)
}
