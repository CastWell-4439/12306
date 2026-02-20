// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Mysql struct {
		Dsn string
	}
	Redis struct {
		Addr     string
		Password string `json:",optional"`
		DB       int    `json:",optional"`
	}
	Kafka struct {
		Brokers []string
	}

	InventoryServiceURL        string `json:",default=http://127.0.0.1:8082"`
	OrderInventoryPartitionKey string `json:",default=G123|2026-02-11|2nd"`
	OrderInventoryDefaultQty   int    `json:",default=1"`
	OrderInventoryCapacity     int    `json:",default=500"`
	PaymentCallbackSignKey     string `json:",optional"`
}
