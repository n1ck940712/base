package db

import (
	"strings"
	"sync"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var (
	useTestDB   = false
	shared      *gorm.DB
	localShared *gorm.DB
	once        sync.Once
)

// set db to nil to force reinit
func ResetConnection() {
	shared = nil
	localShared = nil
}

// use for testing
func UseLocalhost() {
	useTestDB = true
}

func Shared() *gorm.DB {
	if useTestDB {
		once.Do(func() {
			localShared = newDB(localDSN())
		})
		return localShared
	}
	once.Do(func() {
		shared = newDB(settings.GetDBHost().String())
	})
	return shared
}

func newDB(dsn string) *gorm.DB {
	gdb, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
			},
			Logger: logger.Default.LogMode(logger.Silent),
		})

	if err != nil {
		panic("Fail to open DB: " + err.Error())
	}

	//replicas
	if (types.Array[string]{"dev", "live"}).Constains(settings.ENVIRONMENT) && settings.DB_REPLICA != "" {
		gdb.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(settings.DB_REPLICA)},
			Policy:   dbresolver.RandomPolicy{},
		}))
	}
	return gdb
}

func localDSN() string {
	return strings.Replace(settings.DB_HOST, "@db:", "@localhost:", 1)
}
