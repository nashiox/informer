package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "informer"
	app.Version = Version
	app.Usage = ""
	app.Author = "nashiox"
	app.Email = "info@nashio-lab.info"
	app.Commands = Commands

	app.Run(os.Args)
}
