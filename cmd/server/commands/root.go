/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/spf13/cobra"

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
	cfgPathP      string
	configParamsP string

	cfgPath      cmdutil.Flag
	configParams cmdutil.Flag
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
	servicesContext := context.NewServiceContext(cfgPathP)

	log.Root.Info("Run node id: ", servicesContext.Id())

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

func run() {
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
	args := []cmdutil.Flag{cfgPath, configParams}
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
