package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"log"
	"time"

	"github.com/alist-org/alist/v3/cmd/args"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func Log() {
	log.SetOutput(logrus.StandardLogger().Out)
	if args.Debug || args.Dev {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetReportCaller(false)
	}
	logConfig := conf.Conf.Log
	if logConfig.Enable {
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
			logrus.Fatalf("failed to create rotate logrus: %s", err)
		}
		logrus.SetOutput(writer)
	}
	logrus.Infof("init logrus...")
}
