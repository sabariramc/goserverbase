package mysql

import (
	"fmt"
	"log"
	"net/url"

	"sabariram.com/goserverbase/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	conn *gorm.DB
}

func NewConnection(config *config.MySqlConnectionConfig) *DB {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&loc=%v", config.Username, config.Password, config.Host, config.Port, config.DatabaseName, config.Charset, url.QueryEscape(config.Timezone))
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return NewDatabase(conn)
}

func NewDatabase(conn *gorm.DB) *DB {
	return &DB{conn: conn}
}

func (d *DB) GetDB() *gorm.DB {
	return d.conn
}

func (d *DB) Close() {
	sqlDB, err := d.conn.DB()
	if err != nil {
		log.Fatalln(err)
	}
	sqlDB.Close()
}
