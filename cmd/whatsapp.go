package main

import (
	"fmt"
	"os"

	"github.com/piusalfred/whatsapp/cli"
)

func main() {

	if err := cli.NewApp().Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
