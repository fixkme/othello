package main

import (
	"fmt"

	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/othello/server/gate/internal"
)

func main() {
	start()
}

func start() {
	fmt.Println("start gate server")
	app.DefaultApp().Run(
		internal.NewGateModule(),
	)
}
