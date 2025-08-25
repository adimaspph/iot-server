package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewDatabase(viper *viper.Viper, log *logrus.Logger) *sql.DB {
	// Read config using Viper
	username := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASS")
	host := viper.GetString("DB_HOST")
	port := viper.GetInt("DB_PORT")
	database := viper.GetString("DB_NAME")
	idleConnection := viper.GetInt("DB_MAX_IDLE_CON")
	maxConnection := viper.GetInt("DB_MAX_OPEN_CON")
	maxLifeTimeConnection := viper.GetInt("DB_MAX_LIFETIME")
	maxIdleTimeConnection := viper.GetInt("DB_MAX_IDLE_TIME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True", username, password, host, port, database)

	connection, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		panic(err)
	}

	for i := 0; i < 10; i++ {
		err := connection.Ping()
		if err == nil {
			break
		}
		log.Warn("Waiting for database...")
		time.Sleep(2 * time.Second)
	}

	if err := connection.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
		panic(err)
	}

	connection.SetMaxIdleConns(idleConnection)
	connection.SetMaxOpenConns(maxConnection)
	connection.SetConnMaxLifetime(time.Second * time.Duration(maxLifeTimeConnection))
	connection.SetConnMaxIdleTime(time.Second * time.Duration(maxIdleTimeConnection))

	return connection
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}
