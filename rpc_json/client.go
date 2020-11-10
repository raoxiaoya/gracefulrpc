package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"

	"github.com/phprao/gracefulrpc/rpc_json/repo"
)

func main() {
	client, err := jsonrpc.Dial("tcp", "127.0.0.1:8100")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	orderRequest := repo.OrderRequest{OrderId: "asddddddd"}
	var orderInfo repo.OrderInfo

	err = client.Call("Order.GetOne", orderRequest, &orderInfo)
	if err != nil {
		log.Fatal("Order error:", err)
	}
	fmt.Println(orderInfo)
}
