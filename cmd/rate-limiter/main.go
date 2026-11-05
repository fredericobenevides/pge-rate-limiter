package main

import (
	"fmt"
	"net/http"

	"github.com/fredericobenevides/pge-rate-limiter/configs"
	"github.com/fredericobenevides/pge-rate-limiter/internal/handler"
)

func main() {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", handler.HelloWorldHandler)

	var httpPort = fmt.Sprintf(":%s", cfg.Port)

	err = http.ListenAndServe(httpPort, nil)
	if err != nil {
		panic(err)
	}
}
