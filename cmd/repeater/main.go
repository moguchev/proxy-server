package main

import (
	"github.com/moguchev/proxy-server/internal/repeater"
	"log"
	"os"
)

func main() {
	pathToConfig := ""
	if len(os.Args) != 2 {
		panic("Usage: ./main <path_to_config>")
	} else {
		pathToConfig = os.Args[1]
	}
	serv, err := repeater.NewServer(pathToConfig)
	if err != nil {
		panic(err)
	}
	err = serv.Run()
	if err != nil {
		log.Println(err)
	}
}
