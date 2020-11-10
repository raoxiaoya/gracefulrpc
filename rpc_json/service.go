package main

import (
	"log"
	"net/rpc"
	"time"

	"github.com/phprao/gracefulrpc"
	"github.com/phprao/gracefulrpc/rpc_json/repo"
)

func main() {
	err := rpc.Register(new(repo.Order))
	if err != nil {
		log.Fatal(err)
	}

	srv := gracefulrpc.NewServer(gracefulrpc.Config{
		DelayTime: 1*time.Minute,
		CodecType: "json",
	})

	srv.ListenAndServe("tcp", ":8100")
}
