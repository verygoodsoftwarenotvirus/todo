package main

import (
	//

	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
)

func main() {
	cfg, err := config.ParseConfigFile("config_files/production.toml")
	if err != nil {
		log.Fatal(err)
	}

	_ = cfg
	println()
}
