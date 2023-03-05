package bootstrap

import (
	"io"
	"log"
	"os"

	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func init() {
	formatter := logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	}
	logrus.SetFormatter(&formatter)
	utils.Log.SetFormatter(&formatter)
	// logrus.SetLevel(logrus.DebugLevel)
}

func setLog(l *logrus.Logger) {
	if flags.Debug || flags.Dev {
		l.SetLevel(logrus.DebugLevel)
		l.SetReportCaller(true)
	} else {
		l.SetLevel(logrus.InfoLevel)
		l.SetReportCaller(false)
	}
}

func Log() {
	setLog(logrus.StandardLogger())
	setLog(utils.Log)
	logConfig := conf.Conf.Log
	if logConfig.Enable {
		var w io.Writer = &lumberjack.Logger{
			Filename:   logConfig.Name,
			MaxSize:    logConfig.MaxSize, // megabytes
			MaxBackups: logConfig.MaxBackups,
			MaxAge:     logConfig.MaxAge,   //days
			Compress:   logConfig.Compress, // disabled by default
		}
		if flags.Debug || flags.Dev || flags.LogStd {
			w = io.MultiWriter(os.Stdout, w)
		}
		logrus.SetOutput(w)
	}
	log.SetOutput(logrus.StandardLogger().Out)
	utils.Log.Infof("init logrus...")
}
