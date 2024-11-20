package dao

import (
	"context"
	"fmt"
	"gocache/config"
	logger2 "gocache/utils/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strings"
)

var _db *gorm.DB

func InitDB() {
	mConfig := config.Conf.Mysql
	host := mConfig.Host
	port := mConfig.Port
	database := mConfig.Database
	username := mConfig.Username
	password := mConfig.Password
	charset := mConfig.Charset

	dsn := strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", database, "?charset=", charset, "&parseTime=", "true", "&loc=", "Local"}, "")
	err := Database(dsn)
	if err != nil {
		fmt.Println(err)
		logger2.LogrusObj.Error(err)
	}
}

func Database(connectionString string) error {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connectionString,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	_db = db
	migration()
	return err
}
func NewDBClient(ctx context.Context) *gorm.DB {
	db := _db
	return db.WithContext(ctx)
}
