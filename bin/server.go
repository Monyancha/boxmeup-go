package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cjsaylor/boxmeup-go/config"
	"github.com/cjsaylor/boxmeup-go/routing"
)

func main() {
	router := routing.NewRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.Port), router))
}
