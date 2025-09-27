package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	JWT struct {
		Secret string
		TTL    int64 // seconds
	}
}
