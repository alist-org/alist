package cmd

import (
	"fmt"
	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server at the specified address",
	Long: `Start the server at the specified address
the address is defined in config file`,
	Run: func(cmd *cobra.Command, args []string) {
		Init()
		if !flags.Debug && !flags.Dev {
			gin.SetMode(gin.ReleaseMode)
		}
		r := gin.New()
		r.Use(gin.LoggerWithWriter(log.StandardLogger().Out), gin.RecoveryWithWriter(log.StandardLogger().Out))
		server.Init(r)
		base := fmt.Sprintf("%s:%d", conf.Conf.Address, conf.Conf.Port)
		log.Infof("start server @ %s", base)
		var err error
		if conf.Conf.Scheme.Https {
			err = r.RunTLS(base, conf.Conf.Scheme.CertFile, conf.Conf.Scheme.KeyFile)
		} else {
			err = r.Run(base)
		}
		if err != nil {
			log.Errorf("failed to start: %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
