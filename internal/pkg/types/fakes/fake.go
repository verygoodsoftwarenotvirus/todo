package fakes

import (
	"time"

	fake "github.com/brianvoe/gofakeit/v5"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

const (
	exampleQuantity = 3
)
