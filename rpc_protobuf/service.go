package main

import (
	"log"
	"net/rpc"
	"time"

	"gracefulrpc/gracefulrpc"
	"gracefulrpc/rpc_protobuf/repo"
)

func main() {
	err := rpc.Register(new(repo.Order))
	if err != nil {
		log.Fatal(err)
	}

	srv := gracefulrpc.NewServer(gracefulrpc.Config{
		DelayTime: 1*time.Minute,
		CodecType: "protobuf",
	})

	srv.ListenAndServe("tcp", ":8100")
}
