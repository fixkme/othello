package pb

import (
	_ "github.com/fixkme/othello/server/pb/hall"
	_ "github.com/fixkme/othello/server/pb/game"
)

var RequestMsgNames = []string{
	"hall.CLogin",
	"hall.CEnterGame",
	"hall.CLeaveGame",
	"game.CReadyGame",
	"game.CPlacePiece",
}

var ResponseMsgNames = []string{
	"hall.SLogin",
	"hall.SEnterGame",
	"hall.SLeaveGame",
	"game.SReadyGame",
	"game.SPlacePiece",
}

var NoticeMsgNames = []string{
	"hall.PPlayerJoinGame",
	"game.PReadyGame",
	"game.PStartGame",
	"game.PPlacePiece",
	"game.PGameResult",
}

