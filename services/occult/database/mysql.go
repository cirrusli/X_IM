package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"strings"
	"time"
)

func InitDB(driver string, dsn string) (*gorm.DB, error) {
	// dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

	defaultLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      true,
	})

	var dialector gorm.Dialector
	//todo 扩展其他数据库
	if driver == "mysql" {
		dialector = mysql.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: defaultLogger,
		NamingStrategy: schema.NamingStrategy{
			// table name prefix, table for `User` would be `t_users`
			TablePrefix: "t_",
			// use singular table name, table for `User` would be `user` with this option enabled
			SingularTable: true,
			// use name replacer to change struct/field name before convert it to DB name
			NameReplacer: strings.NewReplacer("CID", "Cid"),
		}})

	return db, err
}
