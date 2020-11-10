/*
-- @Time : 2020/11/3 14:41
-- @Author : raoxiaoya
-- @Desc :
*/
package repo

import (
	"errors"
)

type Order struct {}

type OrderInfo struct {
	Id string
	Price float64
	Status int
}

type OrderRequest struct {
	OrderId string
}

func (o *Order) GetOne(orderRequest OrderRequest, orderInfo *OrderInfo) error {
	if orderRequest.OrderId == "" {
		return errors.New("orderId is invalid")
	}

	*orderInfo = OrderInfo{
		Id: orderRequest.OrderId,
		Price: 100.00,
		Status: 1,
	}

	return nil
}
