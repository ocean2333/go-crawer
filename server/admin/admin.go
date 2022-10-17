package main

import (
	"fmt"
	"os"

	"github.com/ocean2333/go-crawer/server/admin/cmds/album"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:              "admin",
		Short:            "admin is a tool to manage go-crawer",
		SilenceErrors:    true,
		SilenceUsage:     true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return fmt.Errorf("unknown go-crawer command: %q", args[0])
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}

	root.AddCommand(
		album.NewCommand(),
	)

	// flags for all commands

	if err := root.Execute(); err != nil {
		fmt.Printf("err: %v", err)
		os.Exit(1)
	}
}
