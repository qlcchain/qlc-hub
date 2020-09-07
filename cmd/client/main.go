/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"os"

	"github.com/qlcchain/qlc-hub/cmd/client/commands"
)

func main() {
	commands.Execute(os.Args)
}
