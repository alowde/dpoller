package logger

import (
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/node"
	"github.com/mattn/go-colorable"
)

// New returns a *logrus.Entry initialised with a few standard settings.
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
