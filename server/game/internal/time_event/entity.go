package time_event

type Global struct {
	EndTime    int64 // 结束时间
	CallbackId int64 // 回调ID
	Event      Event // 定时器事件
}

type Desk struct {
	EndTime    int64 // 结束时间
	CallbackId int64 // 回调ID
	Event      Event // 定时器事件
}
