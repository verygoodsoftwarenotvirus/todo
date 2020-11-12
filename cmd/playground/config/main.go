package main

import (
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"

	//
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/viper"
	"io/ioutil"
)

const exampleConfig = ``

func mustnt(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	f, fileCreateErr := ioutil.TempFile("", "*.toml")
	mustnt(fileCreateErr)

	_, writeErr := f.WriteString(exampleConfig)
	mustnt(writeErr)

	cfg, configParseErr := viper.ParseConfigFile(noop.NewLogger(), f.Name())
	mustnt(configParseErr)

	fmt.Println(cfg.Database.CreateTestUser)
}
