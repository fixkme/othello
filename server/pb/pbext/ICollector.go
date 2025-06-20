package pbext

type ICollector interface {
	Collect(string, any, bool)
}
