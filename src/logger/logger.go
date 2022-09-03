package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	initLogger()
}

func initLogger() {
	Log = logrus.New()
	Log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "2006-01-02T16:04:05.000-0700",
		FieldsOrder:     []string{"component", "category"},
	})

}
