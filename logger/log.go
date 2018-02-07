package logger

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"github.com/mattn/go-colorable"
)

func New(routine string, level logrus.Level) *logrus.Entry {
	var log = &logrus.Logger{
		Out:       colorable.NewColorableStdout(),
		Formatter: &logrus.TextFormatter{ForceColors: true},
		Level:     level,
	}
	return log.WithFields(logrus.Fields{
		"routine": routine,
		"ID":      node.Self.ID,
	})

}
