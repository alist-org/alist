package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alist-org/alist/v3/cmd/flags"
	_ "github.com/alist-org/alist/v3/drivers"
	"github.com/alist-org/alist/v3/internal/bootstrap"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ServerCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server at the specified address",
	Long: `Start the server at the specified address
the address is defined in config file`,
	Run: func(cmd *cobra.Command, args []string) {
		Init()
		if conf.Conf.DelayedStart != 0 {
			utils.Log.Infof("delayed start for %d seconds", conf.Conf.DelayedStart)
			time.Sleep(time.Duration(conf.Conf.DelayedStart) * time.Second)
		}
		bootstrap.InitAria2()
		bootstrap.InitQbittorrent()
		bootstrap.LoadStorages()
		if !flags.Debug && !flags.Dev {
			gin.SetMode(gin.ReleaseMode)
		}
		r := gin.New()
		r.Use(gin.LoggerWithWriter(log.StandardLogger().Out), gin.RecoveryWithWriter(log.StandardLogger().Out))
		server.Init(r)
		var httpSrv, httpsSrv *http.Server
		if !conf.Conf.Scheme.DisableHttp {
			httpBase := fmt.Sprintf("%s:%d", conf.Conf.Address, conf.Conf.Port)
			utils.Log.Infof("start HTTP server @ %s", httpBase)
			httpSrv = &http.Server{Addr: httpBase, Handler: r}
			go func() {
				err := httpSrv.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					utils.Log.Fatalf("failed to start: %s", err.Error())
				}
			}()
		}
		if conf.Conf.Scheme.Https {
			httpsBase := fmt.Sprintf("%s:%d", conf.Conf.Address, conf.Conf.HttpsPort)
			utils.Log.Infof("start HTTPS server @ %s", httpsBase)
			httpsSrv = &http.Server{Addr: httpsBase, Handler: r}
			go func() {
				err := httpsSrv.ListenAndServeTLS(conf.Conf.Scheme.CertFile, conf.Conf.Scheme.KeyFile)
				if err != nil && err != http.ErrServerClosed {
					utils.Log.Fatalf("failed to start: %s", err.Error())
				}
			}()
		}
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 1 second.
		quit := make(chan os.Signal, 1)
		// kill (no param) default send syscanll.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		utils.Log.Println("Shutdown server...")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		if !conf.Conf.Scheme.DisableHttp {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := httpSrv.Shutdown(ctx); err != nil {
					utils.Log.Fatal("HTTP server shutdown:", err)
				}
			}()
		}
		if conf.Conf.Scheme.Https {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := httpsSrv.Shutdown(ctx); err != nil {
					utils.Log.Fatal("HTTPS server shutdown:", err)
				}
			}()
		}
		wg.Wait()
		utils.Log.Println("Server exit")
	},
}

func init() {
	RootCmd.AddCommand(ServerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// OutAlistInit 暴露用于外部启动server的函数
func OutAlistInit() {
	var (
		cmd  *cobra.Command
		args []string
	)
	ServerCmd.Run(cmd, args)
}
