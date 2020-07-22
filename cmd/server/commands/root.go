/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/spf13/cobra"

	sdk "github.com/qlcchain/qlc-go-sdk/pkg/types"

	cmdutil "github.com/qlcchain/qlc-hub/cmd/util"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/services"
	"github.com/qlcchain/qlc-hub/services/context"
)

var (
	shell       *ishell.Shell
	rootCmd     *cobra.Command
	interactive bool
)

var (
	seedP         string
	cfgPathP      string
	configParamsP string

	seed           cmdutil.Flag
	cfgPath        cmdutil.Flag
	configParams   cmdutil.Flag
	maxAccountSize = 100
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(osArgs []string) {
	if len(osArgs) == 2 && osArgs[1] == "-s" {
		interactive = true
	}
	if interactive {
		shell = ishell.NewWithConfig(
			&readline.Config{
				Prompt:      fmt.Sprintf("%c[1;0;32m%s%c[0m", 0x1B, ">> ", 0x1B),
				HistoryFile: "/tmp/readline.tmp",
				//AutoComplete:      completer,
				InterruptPrompt:   "^C",
				EOFPrompt:         "exit",
				HistorySearchFold: true,
				//FuncFilterInputRune: filterInput,
			})
		shell.Println("QLC hub")
		addCommand()
		shell.Run()
	} else {
		rootCmd = &cobra.Command{
			Use:   "ghub",
			Short: "CLI for QLC Hub Server",
			Long:  `QLC Hub is for QLC/ETH Cross-Chain.`,
			Run: func(cmd *cobra.Command, args []string) {
				err := start()
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		rootCmd.PersistentFlags().StringVar(&cfgPathP, "config", "", "config file")
		rootCmd.PersistentFlags().StringVar(&seedP, "seed", "", "seed for accounts")
		rootCmd.PersistentFlags().StringVar(&configParamsP, "configParams", "", "parameter set that needs to be changed")
		addCommand()
		if err := rootCmd.Execute(); err != nil {
			log.Root.Info(err)
			os.Exit(1)
		}
	}
}

func addCommand() {
	if interactive {
		run()
	}
	hubVersion()
}

func start() error {
	var accounts []*sdk.Account
	servicesContext := context.NewServiceContext(cfgPathP)

	log.Root.Info("Run node id: ", servicesContext.Id())

	if len(seedP) > 0 {
		log.Root.Info("run hub SEED mode")
		sByte, _ := hex.DecodeString(seedP)
		tmp, err := seedToAccounts(sByte)
		if err != nil {
			return err
		}
		accounts = append(accounts, tmp...)
	} else {
		log.Root.Info("run hub without account")
	}

	// save accounts to context
	servicesContext.SetAccounts(accounts)
	// start all services by chain context
	err := servicesContext.Init(func() error {
		return services.RegisterServices(servicesContext)
	})
	if err != nil {
		log.Root.Error(err)
		return err
	}

	err = servicesContext.Start()

	if err != nil {
		return err
	}
	trapSignal()
	return nil
}

func trapSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	servicesContext := context.NewServiceContext(cfgPathP)
	err := servicesContext.Stop()
	if err != nil {
		log.Root.Info(err)
	}

	log.Root.Info("hub closed successfully")
}

func seedToAccounts(data []byte) ([]*sdk.Account, error) {
	seed, err := sdk.BytesToSeed(data)
	if err != nil {
		return nil, err
	}
	var accounts []*sdk.Account
	for i := 0; i < maxAccountSize; i++ {
		account, _ := seed.Account(uint32(i))
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func run() {
	seed = cmdutil.Flag{
		Name:  "seed",
		Must:  false,
		Usage: "seed for wallet,if is nil,just run a node",
		Value: "",
	}
	cfgPath = cmdutil.Flag{
		Name:  "config",
		Must:  false,
		Usage: "config file path",
		Value: "",
	}

	configParams = cmdutil.Flag{
		Name:  "configParam",
		Must:  false,
		Usage: "parameter set that needs to be changed",
		Value: "",
	}
	args := []cmdutil.Flag{seed, cfgPath, configParams}
	s := &ishell.Cmd{
		Name:                "run",
		Help:                "start hub server",
		CompleterWithPrefix: cmdutil.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if cmdutil.HelpText(c, args) {
				return
			}
			if err := cmdutil.CheckArgs(c, args); err != nil {
				cmdutil.Warn(err)
				return
			}
			seedP = cmdutil.StringVar(c.Args, seed)
			cfgPathP = cmdutil.StringVar(c.Args, cfgPath)
			configParamsP = cmdutil.StringVar(c.Args, configParams)

			err := start()
			if err != nil {
				cmdutil.Warn(err)
			}
		},
	}
	shell.AddCmd(s)
}
