package conf

import "sort"

type ServerConfig struct {
	Scheme string
	Host   string
	Path   string
}

var ServerList = map[string]ServerConfig{
	"local": {
		Scheme: "ws",
		Host:   "127.0.0.1:7070",
		Path:   "/ws",
	},
}

func GetAllHostName() []string {
	var ret []string
	for k := range ServerList {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}
