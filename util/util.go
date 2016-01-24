package util

import (
    "os"
    "time"
    log "github.com/Sirupsen/logrus"
)

const ApplicationName    = `corona-ui`
const ApplicationSummary = `A Webkit-based GTK window for running desktop web applications`
const ApplicationVersion = `0.0.1`

var StartedAt = time.Now()

func ParseLogLevel(level string) {
    log.SetOutput(os.Stderr)
    log.SetFormatter(&log.TextFormatter{
        ForceColors: true,
    })

    switch level {
    case `info`:
        log.SetLevel(log.InfoLevel)
    case `warn`:
        log.SetLevel(log.WarnLevel)
    case `error`:
        log.SetLevel(log.ErrorLevel)
    case `fatal`:
        log.SetLevel(log.FatalLevel)
    case `quiet`:
        log.SetLevel(log.PanicLevel)
    default:
        log.SetLevel(log.DebugLevel)
    }
}
