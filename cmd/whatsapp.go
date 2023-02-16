package main

import (
	"fmt"
	"os"

	wcli "github.com/piusalfred/whatsapp/cli"
)

func main() {
	app := wcli.New()

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
