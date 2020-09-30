package log

import (
	"log"
	"os"
	"strings"

	"github.com/toolkits/pkg/logger"
)

type Config struct {
	Path     string
	Level    string
	KeepDays uint
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("[ERROR]")
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
}

func Init(config Config) {
	backend, err := logger.NewFileBackend(config.Path)
	if err != nil {
		log.Fatal("create log backend error: %s\n", err)
	}
	backend.SetRotateByHour(true)
	backend.SetKeepHours(config.KeepDays * 24)

	logger.SetLogging(strings.ToUpper(config.Level), backend)
}

func Close() {
	logger.Close()
}