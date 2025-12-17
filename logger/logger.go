package logger

import "fmt"
import "log"
import "os"

type Severity int

const (
	EMERGENCY Severity = iota
	ALERT
	CRITICAL
	ERROR
	WARNING
	NOTICE
	INFO
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

func Emergency(message string) {
	G_loggers[EMERGENCY].Output(2, message)
}
func Alert(message string) {
	G_loggers[ALERT].Output(2, message)
}
func Critical(message string) {
	G_loggers[CRITICAL].Output(2, message)
}
func Error(message string) {
	G_loggers[ERROR].Output(2, message)
}
func Warning(message string) {
	G_loggers[WARNING].Output(2, message)
}
func Notice(message string) {
	G_loggers[NOTICE].Output(2, message)
}
func Info(message string) {
	G_loggers[INFO].Output(2, message)
}
func Debug(message string) {
	G_loggers[DEBUG].Output(2, message)
}
