package main

import (
	"fmt"
	"log"
	"net/rpc"

	"gracefulrpc/rpc_gob/repo"
)

func main() {
	client, err := rpc.Dial("tcp", "127.0.0.1:8100")
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
