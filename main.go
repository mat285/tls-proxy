package main

import (
	"fmt"
	"os"

	"github.com/blend/go-sdk/graceful"
	"github.com/mat285/tls-proxy/proxy"
	"github.com/spf13/cobra"
)

func run() error {
	file := ""
	cmd := &cobra.Command{
		Use:           "proxy",
		Short:         "proxy",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			p, err := proxy.NewProxyFromFile(file)
			if err != nil {
				return err
			}
			return graceful.Shutdown(p)
		},
	}

	cmd.PersistentFlags().StringVar(
		&file,
		"config",
		file,
		"Config file for the proxy",
	)

	return cmd.Execute()
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
