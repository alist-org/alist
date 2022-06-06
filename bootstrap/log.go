package bootstrap

import (
	"github.com/alist-org/alist/v3/cmd/args"
	"github.com/alist-org/alist/v3/conf"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"time"
)

func Log() {
	if args.Debug {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	}
	log.SetFormatter(&log.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	})
	logConfig := conf.Conf.Log
	if !args.Debug && logConfig.Path != "" {
		var (
			writer *rotatelogs.RotateLogs
			err    error
		)
		if logConfig.Name != "" {
			writer, err = rotatelogs.New(
				logConfig.Path,
				rotatelogs.WithLinkName(logConfig.Name),
				rotatelogs.WithRotationCount(logConfig.RotationCount),
				rotatelogs.WithRotationTime(time.Duration(logConfig.RotationTime)*time.Hour),
			)
		} else {
			writer, err = rotatelogs.New(
				logConfig.Path,
				rotatelogs.WithRotationCount(logConfig.RotationCount),
				rotatelogs.WithRotationTime(time.Duration(logConfig.RotationTime)*time.Hour),
			)
		}
		if err != nil {
			log.Fatalf("failed to create rotate log: %s", err)
		}
		log.SetOutput(writer)
	}
	log.Infof("init log...")
}
