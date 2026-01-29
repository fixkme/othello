package system

import (
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/game/internal/entity"
	"google.golang.org/protobuf/proto"
)

type playerSystem struct {
	defaultModule
}

// Player 玩家系统
var Player = new(playerSystem)

var _ systemModule = (*playerSystem)(nil)

func init() {
	Manager.register(Player)
}

// onInit 初始化
func (s *playerSystem) onInit() {

}

// afterInit 初始化后
func (s *playerSystem) afterInit() {

}

// onClose 关闭
func (s *playerSystem) onClose() {

}

func (s *playerSystem) Init(p *entity.Player) (err error) {

	return nil
}

func (s *playerSystem) OnEnterRoom(p *entity.Player, r *entity.Room) {
	// 设置玩家的redis数据

}

func (s *playerSystem) OnLeaveRoom(pid, roomId int64) {
	err := Global.AsyncExec(func() {
		Global.removePlayerFromRoom(pid, roomId)
	})
	if err != nil {
		mlog.Errorf("%d %d OnLeaveRoom Global.AsyncExec failed", roomId, pid)
	}
	mlog.Infof("roomId:%d, playerId:%d OnLeaveRoom", roomId, pid)

	// 删除玩家的redis相关数据
	//redisClient := core.Redis.WriteClient()

}

func (s *playerSystem) NoticePlayer(p *entity.Player, msg proto.Message) error {
	if p.IsRobot {
		return nil
	}

	return NoticePlayer(msg, p)
}
