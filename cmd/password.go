/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/alist-org/alist/v3/internal/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// passwordCmd represents the password command
var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Show admin user's password",
	Run: func(cmd *cobra.Command, args []string) {
		Init()
		admin, err := db.GetAdmin()
		if err != nil {
			log.Errorf("failed get admin user: %+v", err)
		} else {
			log.Infof("admin user's password is: %s", admin.Password)
		}
	},
}

func init() {
	rootCmd.AddCommand(passwordCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// passwordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// passwordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
