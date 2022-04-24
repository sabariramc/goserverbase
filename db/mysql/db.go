package mysql

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sabariramc/goserverbase/config"
	"github.com/sabariramc/goserverbase/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
	log *log.Logger
}

func NewConnection(ctx context.Context, mysqlConfig *config.MySqlConnectionConfig, log log.Logger, sqlLogConfig *glog.Config) *DB {
	connectionString := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&loc=%v", mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.DatabaseName, mysqlConfig.Charset, url.QueryEscape(mysqlConfig.Timezone))
	if sqlLogConfig == nil {
		sqlLogConfig = defaultConfig
	}
	log.Debug(ctx, "NewConnection.connectionString", connectionString)
	conn, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{
		Logger: NewLogger(&log, sqlLogConfig),
	})
	if err != nil {
		log.Emergency(ctx, "mysql.NewConnection", err, err)
	}
	return NewDatabase(conn, &log)
}

func NewDatabase(conn *gorm.DB, log *log.Logger) *DB {
	return &DB{DB: conn, log: log}
}

func (d *DB) Close() {
	sqlDB, err := d.DB.DB()
	if err != nil {
		d.log.Emergency(context.Background(), "mysql.DB.Close", err, err)
	}
	sqlDB.Close()
}
