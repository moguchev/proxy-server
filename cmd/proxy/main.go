package main

import (
	"log"
	"os"

	"github.com/moguchev/proxy-server/internal/proxy"
)

func main() {
	pathToConfig := ""
	if len(os.Args) != 2 {
		log.Fatalln("Usage: ./main <path_to_config>")
	} else {
		pathToConfig = os.Args[1]
	}
	serv, err := proxy.NewServer(pathToConfig)
	if err != nil {
		log.Fatalln(err)
	}
	err = serv.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
