/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// StartCmd represents the start command
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Silent start alist server with `--force-bin-dir`",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func start() {
	initDaemon()
	if pid != -1 {
		_, err := os.FindProcess(pid)
		if err == nil {
			log.Info("alist already started, pid ", pid)
			return
		}
	}
	args := os.Args
	args[1] = "server"
	args = append(args, "--force-bin-dir")
	cmd := &exec.Cmd{
		Path: args[0],
		Args: args,
		Env:  os.Environ(),
	}
	stdout, err := os.OpenFile(filepath.Join(filepath.Dir(pidFile), "start.log"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(os.Getpid(), ": failed to open start log file:", err)
	}
	cmd.Stderr = stdout
	cmd.Stdout = stdout
	err = cmd.Start()
	if err != nil {
		log.Fatal("failed to start children process: ", err)
	}
	log.Infof("success start pid: %d", cmd.Process.Pid)
	err = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0666)
	if err != nil {
		log.Warn("failed to record pid, you may not be able to stop the program with `./alist stop`")
	}
}

func init() {
	RootCmd.AddCommand(StartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
