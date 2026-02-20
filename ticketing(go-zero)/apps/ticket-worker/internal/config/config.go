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

	SeatAllocatorMode       string `json:",default=mock"`
	SeatAllocatorAddr       string `json:",default=127.0.0.1:50051"`
	SeatAllocatorTrainID    string `json:",default=G123"`
	SeatAllocatorTravelDate string `json:",default=2026-02-11"`
	SeatAllocatorCoachType  string `json:",default=2nd"`
	SeatAllocatorFromIndex  int    `json:",default=1"`
	SeatAllocatorToIndex    int    `json:",default=3"`
}

