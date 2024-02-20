package main

import (
	"context"
	"fmt"

	"github.com/xmcontinue/solana-go"

	"git.cplus.link/crema/backend/chain/sol"
)

func main() {
	cli := sol.NewRPC("https://rpc.ankr.com/solana/b42ba44de37bc8cbf595b13a223a3ce1ad98f474836b2958c9971773d76cf106")
	res, err := cli.GetAccountInfo(context.Background(), solana.MustPublicKeyFromBase58("7MPnn7k6uYSMcW7VvB8GUcEedLApmsAG1Fesfufv7rBk"))
	fmt.Println(res, err)
}
