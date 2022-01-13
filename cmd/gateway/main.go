package main

import (
	"git.cplus.link/go/akit/config"

	"git.cplus.link/crema/backend/cmd/gateway/app"
)

func main() {
	conf := config.NewConfiger()
	if err := app.Start(conf); err != nil {
		panic(err)
	}
}
