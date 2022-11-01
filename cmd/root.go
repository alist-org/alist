package cmd

import (
	"fmt"
	"os"

	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "alist",
	Short: "A file list program that supports multiple storage.",
	Long: `A file list program that supports multiple storage,
built with love by Xhofe and friends in Go/Solid.js.
Complete documentation is available at https://alist.nn.ci/`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flags.DataDir, "data", "data", "config file")
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "start with debug mode")
	rootCmd.PersistentFlags().BoolVar(&flags.NoPrefix, "no-prefix", false, "disable env prefix")
	rootCmd.PersistentFlags().BoolVar(&flags.Dev, "dev", false, "start with dev mode")
	rootCmd.PersistentFlags().BoolVar(&flags.ForceBinDir, "force-bin-dir", false, "Force to use the directory where the binary file is located as data directory")
}
