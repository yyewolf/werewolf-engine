package main

import (
	"fmt"
	"os"

	"github.com/yyewolf/werewolf-engine/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "werewolf-cli: %v\n", err)
		os.Exit(1)
	}
}
