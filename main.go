package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "informer"
	app.Version = Version
	app.Usage = ""
	app.Author = "nashiox"
	app.Email = "u.4.o.holly12@gmail.com"
	app.Commands = Commands

	app.Run(os.Args)
}
