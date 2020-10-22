package main

import (
	"fmt"
	//
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config/viper"
	"io/ioutil"
)

const exampleConfig = `
[auth]
  cookie_lifetime = "24h0m0s"
  cookie_secret = "HEREISA32CHARSECRETWHICHISMADEUP"
  enable_user_signup = true

[database]
  connection_details = "/tmp/db"
  debug = false
  provider = "sqlite"
  run_migrations = true
  should_migrate0 = true

  [database.create_test_user]
    is_admin = true
    password = "integration-tests-are-cool"
    username = "exampleUser"

[frontend]
  static_files_directory = "/frontend"

[meta]
  debug = false
  run_mode = "testing"
  startup_deadline = "1m0s"

[metrics]
  database_metrics_collection_interval = "2s"
  metrics_provider = "prometheus"
  runtime_metrics_collection_interval = "2s"
  tracing_provider = "jaeger"

[search]
  items_index_path = "items.bleve"

[server]
  debug = true
  http_port = 8888
`

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

	cfg, configParseErr := viper.ParseConfigFile(f.Name())
	mustnt(configParseErr)

	fmt.Println(cfg.Database.CreateTestUser)
}
