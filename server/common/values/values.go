package values

// rpc meta的key常量定义
const (
	Rpc_PlayerId     = "player_id"
	Rpc_GateId       = "gate_id"       // 消息上下文的gate id，用于推送通知类型消息
	Rpc_GameId       = "game_id"       // 玩家所在的游戏id
	Rpc_NoticeOffset = "notice_offset" // 对gate的回应，payload包含rspData和noticeData
)

// 服务类型
const (
	Service_Gate = "gate"
	Service_Hall = "hall"
	Service_Game = "game"
)

type LogicContextKeyType string

const (
	RpcContext           LogicContextKeyType = "RpcContext"
	RpcContext_Meta      LogicContextKeyType = "RpcContextMeta"
	RpcContext_RoomAgent LogicContextKeyType = "RpcContextRoomAgent"
)
