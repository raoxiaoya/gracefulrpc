package main

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/mars9/codec"
	"github.com/phprao/gracefulrpc/rpc_protobuf/pbs"
)

func main() {
	Client()
}

func Client() {
	conn, err := net.Dial("tcp", "127.0.0.1:8100")
	if err != nil {
		panic(err)
	}
	client := rpc.NewClientWithCodec(codec.NewClientCodec(conn))

	request := message.OrderRequest{OrderId: "201907310001"}
	var response message.OrderInfo
	err = client.Call("Order.GetOne", &request, &response)
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}