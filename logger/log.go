package log

import (
	"log"
	"os"
	"strings"

	"github.com/toolkits/pkg/logger"
)

// Config log config
type Config struct {
	Path     string // log file dir
	Level    string // log level
	KeepDays uint   // log rotate keep days
}

const (
	defaultLogFlags = log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds
)

// StdoutLogger is default logger print to stdout
var StdoutLogger *log.Logger = log.New(os.Stdout, "[LOG OUT]", defaultLogFlags)

// StderrLogger is default logger print to stdout
var StderrLogger *log.Logger = log.New(os.Stderr, "[LOG ERR]", defaultLogFlags)

// Init log by config
func Init(config Config) {
	backend, err := logger.NewFileBackend(config.Path)
	if err != nil {
		log.Fatalf("create log backend error: %s\n", err.Error())
	}
	backend.SetRotateByHour(true)
	backend.SetKeepHours(config.KeepDays * 24)

	logger.SetLogging(strings.ToUpper(config.Level), backend)
}

// InitDefault changes default log
func InitDefault() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[ERROR]")
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
}

// Close log
func Close() {
	logger.Close()
}
