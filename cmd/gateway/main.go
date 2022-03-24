package main

import (
	"git.cplus.link/go/akit/config"

	"git.cplus.link/crema/backend/cmd/gateway/app"

	_ "git.cplus.link/crema/backend/pkg/errcode"
)

func main() {
	conf := config.NewConfiger()
	if err := app.Start(conf); err != nil {
		panic(err)
	}
}
