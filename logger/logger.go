package logger

import (
	// "errors"
	"fmt"
	"log"
	"os"
)

type Severity int

const (
	// System is unusable.
	EMERGENCY Severity = iota
	// Action must be taken immediately.
	ALERT
	// Critical conditions.
	CRITICAL
	// Error conditions.
	ERROR
	// Warning condition.
	WARNING
	// Conditions that are not error conditions, but that
	// may require special handling.
	NOTICE
	// Informational messages.
	// Confirmation that the program is working as expected.
	INFO
	// Messages that contain information normally of use
	// only when debugging a program.
	DEBUG
)

var G_severity_name = map[Severity]string{
	EMERGENCY: "EMERGENCY",
	ALERT:     "ALERT",
	CRITICAL:  "CRITICAL",
	ERROR:     "ERROR",
	WARNING:   "WARNING",
	NOTICE:    "NOTICE",
	INFO:      "INFO",
	DEBUG:     "DEBUG",
}

var G_loggers = []*log.Logger{
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[EMERGENCY]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[ALERT]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[CRITICAL]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[ERROR]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[WARNING]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[NOTICE]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[INFO]), log.Ltime),
	log.New(os.Stdout, fmt.Sprintf("%s ", G_severity_name[DEBUG]), log.Ltime),
}

func doLog(instance *log.Logger, message string, errs ...error) {
	if nil == instance {
		return
	}

	if len(errs) > 0 {
		for _, err := range errs {
			instance.Output(2, message+err.Error())
		}
	} else {
		instance.Output(2, message)
	}
}

// System is unusable.
func Emergency(message string, errs ...error) {
	doLog(G_loggers[EMERGENCY], message, errs...)
}

// Action must be taken immediately.
func Alert(message string, errs ...error) {
	doLog(G_loggers[ALERT], message, errs...)
}

// Critical conditions.
func Critical(message string, errs ...error) {
	doLog(G_loggers[CRITICAL], message, errs...)
}

// Error conditions.
func Error(message string, errs ...error) {
	doLog(G_loggers[ERROR], message, errs...)
}

// Warning condition.
func Warning(message string, errs ...error) {
	doLog(G_loggers[WARNING], message, errs...)
}

// Conditions that are not error conditions, but that may require special handling.
func Notice(message string, errs ...error) {
	doLog(G_loggers[NOTICE], message, errs...)
}

// Informational messages.
// Confirmation that the program is working as expected.
func Info(message string, errs ...error) {
	doLog(G_loggers[INFO], message, errs...)
}

// Messages that contain information normally of use only when debugging a program.
func Debug(message string, errs ...error) {
	doLog(G_loggers[DEBUG], message, errs...)
}

func Log(severity Severity, message string, errs ...error) {
	switch severity {
	case EMERGENCY:
		Emergency(message, errs...)
	case ALERT:
		Alert(message, errs...)
	case CRITICAL:
		Critical(message, errs...)
	case ERROR:
		Error(message, errs...)
	case WARNING:
		Warning(message, errs...)
	case NOTICE:
		Notice(message, errs...)
	case INFO:
		Info(message, errs...)
	case DEBUG:
		Debug(message, errs...)
	}
}
