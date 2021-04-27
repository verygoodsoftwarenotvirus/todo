package fakes

import fake "github.com/brianvoe/gofakeit/v5"

// BuildFakeTime builds a fake time.
func BuildFakeTime() uint64 {
	return fake.Uint64()
}
