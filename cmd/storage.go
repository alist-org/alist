/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/spf13/cobra"
)

// storageCmd represents the storage command
var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage",
}

func init() {
	var mountPath string
	var disable = &cobra.Command{
		Use:   "disable",
		Short: "Disable a storage",
		Run: func(cmd *cobra.Command, args []string) {
			Init()
			storage, err := db.GetStorageByMountPath(mountPath)
			if err != nil {
				utils.Log.Errorf("failed to query storage: %+v", err)
			} else {
				storage.Disabled = true
				err = db.UpdateStorage(storage)
				if err != nil {
					utils.Log.Errorf("failed to update storage: %+v", err)
				} else {
					utils.Log.Infof("Storage with mount path [%s] have been disabled", mountPath)
				}
			}
		},
	}
	disable.Flags().StringVarP(&mountPath, "mount-path", "m", "", "The mountPath of storage")
	RootCmd.AddCommand(storageCmd)
	storageCmd.AddCommand(disable)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
