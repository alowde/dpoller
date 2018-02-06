package logger

import (
	"github.com/Sirupsen/logrus"
	"github.com/mattn/go-colorable"
)

func New(routine string, level logrus.Level) *logrus.Entry {
	var log = &logrus.Logger{
		Out:       colorable.NewColorableStdout(),
		Formatter: &logrus.TextFormatter{ForceColors: true},
		Level:     level,
	}
	return log.WithField("routine", routine)
}
