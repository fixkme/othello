package system

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/fixkme/gokit/db/mongo/delta"
	"github.com/fixkme/gokit/framework/core"
	g "github.com/fixkme/gokit/framework/go"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/common"
	"github.com/fixkme/othello/server/hall/internal/conf"
	"github.com/fixkme/othello/server/hall/internal/entity"
	"github.com/fixkme/othello/server/hall/internal/logic"
	"github.com/fixkme/othello/server/hall/internal/time_event"
	"github.com/fixkme/othello/server/pb/datas"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/fixkme/othello/server/pb/hall"
	"github.com/fixkme/othello/server/pb/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type globalSystem struct {
	defaultModule
	logicRoutine *g.RoutineAgent

	accPlayers   map[string]int64
	players      map[int64]*entity.Player // playerId -> Player
	tables       map[int64]int64          // tableId -> gameId
	matchPlayers map[int64]int64          // pid => tableId
	matching     *redblacktree.Tree       // 等待匹配

	timerGenId     atomic.Int64
	timerCallbacks map[string]globalTimerCallback
	runningTimers  map[int64]*time_event.Global
	callbackTimers map[int64]int64

	playerMonitor *delta.DeltaMonitor[int64]
	rootCtx       context.Context
	rootCancel    context.CancelFunc
}

type globalTimerCallback func(data time_event.Event, now int64)

// Global 全局系统
var Global = new(globalSystem)

func init() {
	Manager.register(Global)
}

// onInit 初始化
func (s *globalSystem) onInit() {
	s.timerCallbacks = make(map[string]globalTimerCallback)
	s.runningTimers = make(map[int64]*time_event.Global)
	s.callbackTimers = make(map[int64]int64)
	s.accPlayers = make(map[string]int64)
	s.players = make(map[int64]*entity.Player)
	s.tables = make(map[int64]int64)
	s.matchPlayers = make(map[int64]int64)
	s.matching = redblacktree.NewWith(utils.Int64Comparator)
	s.playerMonitor = delta.NewDeltaMonitor[int64](core.Mongo.Client(), conf.DBName, conf.CollPlayer)
	s.rootCtx, s.rootCancel = context.WithCancel(context.Background())
}

// afterInit 初始化后
func (s *globalSystem) afterInit() {
	s.logicRoutine = logic.GetGlobalLogic()
	s.playerMonitor.Start(s.rootCtx)
	go s.runTicker(s.rootCtx)
	// 注册定时器回调

	// 其他全局数据的加载
	if err := s.loadAllPlayer(); err != nil {
		mlog.Fatalf("loadAllPlayer err:%v", err)
	}
	// 其他全局模块的初始化

	// 完全初始化后再执行

}

// Close 关闭
func (s *globalSystem) Close() {
	mlog.Infof("global system closing")
	s.rootCancel()
}

// onClose 关闭
func (s *globalSystem) onClose() {
	s.playerMonitor.Stop()
}

// 所有timer实例的id来源，包括gloabl timer和player timer
func (s *globalSystem) generateTimerId() int64 {
	return s.timerGenId.Add(1)
}

func (s *globalSystem) SyncCall(fn func() error) (err error) {
	fail := s.logicRoutine.SyncRunFunc(func() {
		err = fn()
	})
	if fail != nil {
		return fail
	}
	return err
}

// SyncExec 同步执行
func (s *globalSystem) SyncExec(fn func()) error {
	return s.logicRoutine.SyncRunFunc(fn)
}

// AsyncExec 异步执行
func (s *globalSystem) AsyncExec(fn func()) error {
	return s.logicRoutine.TryRunFunc(fn)
}

// MetricsCollect 收集监控指标
func (s *globalSystem) MetricsCollect() {
}

// runTicker 运行定时器
func (s *globalSystem) runTicker(ctx context.Context) {
	// 监控统计
	const metricsDuration = 5 * time.Second
	metricsTicker := time.NewTicker(metricsDuration)
	defer metricsTicker.Stop()

	const savePlayerDuration = 15 * time.Second
	savePlayeTicker := time.NewTicker(savePlayerDuration)
	defer savePlayeTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-metricsTicker.C:
			s.AsyncExec(s.MetricsCollect)
		case <-savePlayeTicker.C:
			s.AsyncExec(s.playerMonitor.SaveChangedDatas)
		}
	}
}

// TODO 优化，只在登录时加载玩家数据， 这里只加载所有 account => playerId
func (s *globalSystem) loadAllPlayer() error {
	coll := core.Mongo.Client().Database(conf.DBName).Collection(conf.CollPlayer)
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
		p := entity.NewPlayer(pbData.PlayerId, pbData)
		s.players[pbData.PlayerId] = p
		s.accPlayers[pbData.Account] = pbData.PlayerId
	}
	mlog.Infof("load player data finished, size:%d", len(s.players))
	return nil
}

type IdSeq struct {
	Id  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

func (s *globalSystem) GeneId(idName string) (id int64, err error) {
	coll := core.Mongo.Client().Database(conf.DBName).Collection(conf.IdCollName)
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

func (s *globalSystem) GetPlayerByAccount(acc string) *entity.Player {
	id, ok := s.accPlayers[acc]
	if !ok {
		return nil
	}
	p := s.GetPlayer(id)
	return p
}

func (s *globalSystem) GetPlayer(playerId int64) *entity.Player {
	return s.players[playerId]
}

func (s *globalSystem) GetTableGameNodeId(tableId int64) int64 {
	return s.tables[tableId]
}

func (s *globalSystem) CreatePlayer(acc string, pinfo *datas.PBPlayerInfo) (*entity.Player, error) {
	pid, ex := s.accPlayers[acc]
	if !ex {
		id, err := s.GeneId(conf.IdName_Player)
		if err != nil {
			return nil, err
		}
		pid = id
	}

	p := entity.NewPlayer(pid, nil)
	p.BindMonitor(s.playerMonitor, p)
	p.SetPlayerId(pid)
	p.SetAccount(acc)
	p.GetModelPlayerInfo().InitFromPB(pinfo)
	Player.InitModels(p)
	s.players[pid] = p
	s.accPlayers[acc] = pid
	s.playerMonitor.SaveEntity(p)
	return p, nil
}

func (s *globalSystem) RemovePlayer(playerId int64) {
	_, ok := s.players[playerId]
	if !ok {
		return
	}
	delete(s.players, playerId)
}

func (s *globalSystem) RemoveTable(tid int64) {
	_, ok := s.tables[tid]
	if !ok {
		return
	}
	delete(s.tables, tid)
}

func (s *globalSystem) CreateMatchTable(p *entity.Player) (*datas.PBTableLocation, error) {
	tid, err := s.GeneId(conf.IdName_Table)
	if err != nil {
		return nil, err
	}
	tb := datas.NewPBTableLocation()
	tb.TableId = tid
	tb.PlayType = datas.PlayType_PT_Common
	tb.GameId = 1
	// TODO  负载均衡选择game节点，现在默认game节点为1
	s.tables[tid] = tb.GameId

	tb.Player1 = p.Id()
	s.matching.Put(tb.GetTableId(), tb)
	s.matchPlayers[p.Id()] = tid
	return tb, nil
}

func (s *globalSystem) GetMatchingTable(pid int64) *datas.PBTableLocation {
	tid := s.matchPlayers[pid]
	if tid == 0 {
		return nil
	}
	v, ok := s.matching.Get(tid)
	if !ok {
		return nil
	}
	tb := v.(*datas.PBTableLocation)
	return tb
}

func (s *globalSystem) GetMatchTable() *datas.PBTableLocation {
	node := s.matching.Left()
	if node == nil {
		return nil
	}
	tb := node.Value.(*datas.PBTableLocation)
	return tb
}

func (s *globalSystem) PlayerMatchSucceed(p1, p2 *entity.Player, tb *datas.PBTableLocation) {
	tb.Player2 = p2.Id()
	tb.CreateTime = util.NowMs()
	s.matching.Remove(tb.GetTableId())
	delete(s.matchPlayers, p1.Id())
	delete(s.matchPlayers, p2.Id())

	loc := datas.NewMTableLocation()
	loc.InitFromPB(tb)
	p1.SetInTables(int64(tb.GetPlayType()), loc)

	loc = datas.NewMTableLocation()
	loc.InitFromPB(tb)
	p2.SetInTables(int64(tb.GetPlayType()), loc)
}

func (s *globalSystem) PlayerLeaveGame(p *entity.Player) error {
	s.RemoveMatchPlayer(p)

	if mtb, _ := p.GetInTables(int64(datas.PlayType_PT_Common)); mtb != nil { //已开局
		gameService := GameService(mtb.GetGameId())
		req := &game.CLeaveGame{PlayerId: p.Id()}
		resp := &game.SLeaveGame{}
		err := core.Rpc.SyncCall(gameService, req, resp, 0, common.WarpMeta(p.Id(), p.GateId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *globalSystem) RemoveMatchPlayer(p *entity.Player) bool {
	if tid := s.matchPlayers[p.Id()]; tid > 0 {
		delete(s.matchPlayers, p.Id())
		// 移除匹配
		if v, ok := s.matching.Get(tid); ok && v != nil {
			s.matching.Remove(tid)
			s.RemoveTable(tid)
			return true
		}
	}
	return false
}

func (s *globalSystem) GameSettle(in *hall.CGameSettle) {
	pt := int64(datas.PlayType_PT_Common)
	if p := s.GetPlayer(in.OwnerPlayer.Id); p != nil {
		p.RemoveInTables(pt)
	}
	if p := s.GetPlayer(in.OppoPlayer.Id); p != nil {
		p.RemoveInTables(pt)
	}

	tid := in.TableId
	s.matching.Remove(tid) //保险起见
	s.RemoveTable(tid)
}
