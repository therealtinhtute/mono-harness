package main

import (
	"os"

	"github.com/therealtinhtute/skills/cli/internal/interfaces"
)

var version = "dev"

func main() {
	os.Exit(interfaces.Execute(version))
}
