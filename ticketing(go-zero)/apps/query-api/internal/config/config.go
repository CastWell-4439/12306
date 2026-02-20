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
}
