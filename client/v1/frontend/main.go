// +build wasm

package main

import (
	"fmt"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/verso/html"
)

func main() {
	fmt.Println("getting body")
	body := html.Body()

	button := html.NewButton("Button")
	button.OnClick(func() {
		html.Alert("button clicked!")
	}, false)

	body.AppendChild(button)

	// suspend loop
	for {
		select {
		case <-time.NewTicker((1<<31 - 1) * time.Second).C:
			//
		}
	}
}
