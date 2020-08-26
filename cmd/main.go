package main

import (
	"fmt"
	"os"

	version "github.com/qlcchain/qlc-hub"
	server "github.com/qlcchain/qlc-hub/cmd/server/commands"
)

func main() {
	fmt.Println(version.ShortVersion())
	server.Execute(os.Args)
}
