package models

type Order struct {
	Id       int        `json:"id"`
	Status   string     `json:"status"`
	MetaData []MetaData `json:"meta_data"`
}

type MetaData struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type OrderStatus struct {
	Status         int    `json:"status"`
	DeliveryStatus string `json:"delivery_status"`
}

type ChangeStatus struct {
	Status string `json:"status"`
}

type Balance struct {
	Status int `json:"status"`
}
