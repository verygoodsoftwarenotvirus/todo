package fakes

import (
	"fmt"
	"time"

	fake "github.com/brianvoe/gofakeit/v5"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

const (
	exampleQuantity = 3
)

func BuildFakeSQLQuery() (string, []interface{}) {
	s := fmt.Sprintf("%s %s WHERE things = ? AND stuff = ?",
		fake.RandString([]string{"SELECT * FROM", "INSERT INTO", "UPDATE"}),
		fake.Word(),
	)

	return s, []interface{}{"things", "stuff"}
}
