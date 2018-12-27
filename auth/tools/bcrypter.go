package main

import (
	"fmt"
	"log"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("no input provided")
	}

	if newPass, err := auth.NewBcrypt(nil).HashPassword(os.Args[1]); err != nil {
		log.Fatalf("error hashing password: %v", err)
	} else {
		fmt.Println(newPass)
	}
}
