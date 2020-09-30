package storage

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	rlogs "github.com/lestrrat-go/file-rotatelogs"
)

type Config struct {
	Type      string
	WriterDSN string
	ReaderDSN string

	LogPath             string
	LogLevel            string
	LogMaxDays          uint
	LogRotationHours    uint
	LogSlowMicroSeconds uint

	ConnMaxLifeSeconds   uint
	TimeoutMilliseSecond uint
	MaxOpenConns         int
	MaxIdleConns         int
}

// TODO: move to tooltips repo
func InitDatabase(config Config) *sql.DB {
	// 初始化数据库日志
	logWriter, err := rlogs.New(
		config.LogPath,
		rlogs.WithMaxAge(time.Duration(config.LogMaxDays)*24*time.Hour),
		rlogs.WithRotationTime(time.Duration(config.LogRotationHours)*time.Hour),
	)
	if err != nil {
		log.Fatalf("create db log failed! Error: %s\n", err)
	}
	ormLoggerConfig := logger.Config{
		SlowThreshold: time.Microsecond,
		LogLevel: func(level string) logger.LogLevel {
			switch strings.ToLower(level) {
			case "silent":
				return logger.Silent
			case "error":
				return logger.Error
			case "warn":
				return logger.Warn
			case "info":
				return logger.Info
			case "debug":
				return logger.Info
			}
			return logger.Error
		}(config.LogLevel),
	}
	ormlogger := logger.New(
		log.New(logWriter, "[gorm]", log.LstdFlags),
		ormLoggerConfig,
	)

	// 创建数据库链接
	var conn gorm.Dialector
	switch config.Type {
	case "mysql":
		conn = mysql.Open(config.WriterDSN)
	case "sqlite":
		conn = sqlite.Open(config.WriterDSN)
		config.ReaderDSN = ""
	default:
		log.Fatalf("storage type not supported! Type: %s\n", config.Type)
	}

	// 启动数据库链接
	db, err := gorm.Open(conn, &gorm.Config{
		Logger: ormlogger,
	})

	// 设置读写分离
	if len(config.ReaderDSN) > 0 {
		db.Use(dbresolver.Register(
			dbresolver.Config{
				Replicas: []gorm.Dialector{mysql.Open((config.ReaderDSN))},
			},
		))
	}

	// 设置链接池
	sqldb, err := db.DB()
	sqldb.SetMaxIdleConns(config.MaxIdleConns)
	sqldb.SetMaxOpenConns(config.MaxOpenConns)
	sqldb.SetConnMaxLifetime(time.Duration(config.ConnMaxLifeSeconds) * time.Second)

	// 测试链接
	if err := sqldb.Ping(); err != nil {
		log.Fatalf("connect to database failed when ping the server! Error: %s\n", err)
	}

	return sqldb
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Close database error: %s\n", err.Error())
	}
}
