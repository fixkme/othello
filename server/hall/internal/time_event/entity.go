package time_event

type Global struct {
	EndTime    int64 // 结束时间
	CallbackId int64 // 回调ID
	Event      Event // 定时器事件
}

type Player struct {
	TimerId    int64 // 定时器ID
	Id         int64 // 玩家ID
	EndTime    int64 // 结束时间
	CallbackId int64 // 回调ID
	Event      Event // 定时器事件
}
