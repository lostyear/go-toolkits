package storage

import (
	"log"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	rlogs "github.com/lestrrat-go/file-rotatelogs"
)

// Config storage config
type Config struct {
	Type      string // storage type, sqlite or mysql
	WriterDSN string // storage writer
	ReaderDSN string // storage reader

	TablePrefix   string // prefix for tables
	SingularTable bool   // is tablename has -s suffix

	LogPath             string // db log file path
	LogLevel            string // db log level
	LogMaxDays          uint   // db log keep days
	LogRotationHours    uint   // db log rotate time
	LogSlowMicroSeconds uint   // db slow log time

	ConnMaxLifeSeconds   uint // db connection max keep time
	TimeoutMilliseSecond uint // db request timeout
	MaxOpenConns         int  // max db connections
	MaxIdleConns         int  // free db connections
}

// InitDatabase init db engine by config
func InitDatabase(config Config) *gorm.DB {
	// 初始化数据库日志
	logWriter, err := rlogs.New(
		config.LogPath+".%Y%m%d%H",
		rlogs.WithRotationTime(time.Duration(config.LogRotationHours)*time.Hour),
		rlogs.WithMaxAge(time.Duration(config.LogMaxDays)*24*time.Hour),
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
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   config.TablePrefix,
			SingularTable: config.SingularTable,
		},
	})
	if err != nil {
		log.Fatalf("open db connection failed! Error: %s\n", err)
	}

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
	if err != nil {
		log.Fatalf("get db sql connection failed! Error: %s\n", err)
	}
	sqldb.SetMaxIdleConns(config.MaxIdleConns)
	sqldb.SetMaxOpenConns(config.MaxOpenConns)
	sqldb.SetConnMaxLifetime(time.Duration(config.ConnMaxLifeSeconds) * time.Second)

	// 测试链接
	if err := sqldb.Ping(); err != nil {
		log.Fatalf("connect to database failed when ping the server! Error: %s\n", err)
	}

	return db
}

// Close db
func Close(db *gorm.DB) {
	sqldb, err := db.DB()
	if err != nil {
		log.Printf("get db sql connection failed! Error: %s\n", err)
	}
	if err := sqldb.Close(); err != nil {
		log.Printf("Close database error: %s\n", err.Error())
	}
}
