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

	InventoryShardCount           int   `json:",default=32"`
	InventoryWALBuffer            int   `json:",default=4096"`
	InventorySnapshotIntervalSecs int   `json:",default=10"`
	InventorySnapshotOpsThreshold int64 `json:",default=500"`
	InventoryHoldTTLSecs          int   `json:",default=120"`
}
