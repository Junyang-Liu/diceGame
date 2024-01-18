package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func init() {
	Logger.Out = os.Stdout
	// Logger.Formatter = &logrus.JSONFormatter{}
	Logger.SetReportCaller(true)
	Logger.SetLevel(logrus.TraceLevel)
}

func SetLogLevel(lvl string) {
	l, er := logrus.ParseLevel(lvl)
	if er != nil {
		Logger.Error(er.Error())
	} else {
		Logger.SetLevel(l)
	}
}

func New(prefix string) {

}
