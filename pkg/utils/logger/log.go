package logger

import "github.com/sirupsen/logrus"

var LogrusObj *logrus.Logger

func InitLog() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	LogrusObj = log
}
