/*
-- @Time : 2020/11/3 14:41
-- @Author : raoxiaoya
-- @Desc :
*/
package repo

import (
	"errors"

	message "github.com/phprao/gracefulrpc/rpc_protobuf/pbs"
)

type Order struct {}

func (o *Order) GetOne(orderRequest *message.OrderRequest, orderInfo *message.OrderInfo) error {
	if orderRequest.OrderId == "" {
		return errors.New("orderId is invalid")
	}

	*orderInfo = message.OrderInfo{
		Id: orderRequest.OrderId,
		Price: 100.00,
		Status: 1,
	}

	return nil
}
