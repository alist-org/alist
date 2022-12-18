/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// RestartCmd represents the restart command
var RestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart alist server by daemon/pid file",
	Run: func(cmd *cobra.Command, args []string) {
		stop()
		start()
	},
}

func init() {
	RootCmd.AddCommand(RestartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restartCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restartCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
