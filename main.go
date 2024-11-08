package main

import (
	"embed"
	"log"

	"github.com/usesend0/send0/cmd"
)

var migrations embed.FS
var version string = "dev"

func main() {
	err := cmd.Execute(version, migrations)
	if err != nil {
		log.Fatal(err)
	}
}
