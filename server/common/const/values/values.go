package values

// rpc meta的key常量定义
const (
	Rpc_SessionId    = "session_id"
	Rpc_GateId       = "gate_id"       // 消息上下文的gate id，用于推送通知类型消息
	Rpc_NoticeOffset = "notice_offset" // 对gate的回应，payload包含rspData和noticeData
)

// 服务类型
const (
	Service_Gate = "gate"
	Service_Game = "game"
)
