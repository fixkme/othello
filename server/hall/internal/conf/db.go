package conf

import "time"

const (
	DBName     = "othello_db"
	CollPlayer = "player"

	IdCollName    = "id_gen"
	IdName_Player = "player_id_gen"
	IdName_Table  = "table_id_gen"

	MongoTimeout = 10 * time.Second
)
