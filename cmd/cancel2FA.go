/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/spf13/cobra"
)

// Cancel2FACmd represents the delete2fa command
var Cancel2FACmd = &cobra.Command{
	Use:   "cancel2fa",
	Short: "Delete 2FA of admin user",
	Run: func(cmd *cobra.Command, args []string) {
		Init()
		defer Release()
		admin, err := op.GetAdmin()
		if err != nil {
			utils.Log.Errorf("failed to get admin user: %+v", err)
		} else {
			err := op.Cancel2FAByUser(admin)
			if err != nil {
				utils.Log.Errorf("failed to cancel 2FA: %+v", err)
			} else {
				utils.Log.Info("2FA canceled")
				DelAdminCacheOnline()
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(Cancel2FACmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cancel2FACmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cancel2FACmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
