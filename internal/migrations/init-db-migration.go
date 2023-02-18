package migrations

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	initDB = Init()
	DB     = New(initDB).DB
)

type handler struct {
	DB *gorm.DB
}

func Init() *gorm.DB {
	_ = godotenv.Load(".env")
	dbURL := settings.DB_HOST

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "mini_game_",
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
	})

	if err != nil {
		panic("Fail to open DB: " + err.Error())
	}

	return db
}

func New(db *gorm.DB) handler {
	return handler{db}
}

type Migrations struct {
	indexes []string
}

type IMigrations interface {
	Migrate()
}

func NewMigrations() *Migrations {
	return &Migrations{}
}

func (m *Migrations) Migrate() {

	logger.Debug("------------ DB migrations Status [Initiating] ------------")
	m.migrateLolTowerMemberLevel()
	m.migrateComboTicket()
	m.MigrateMiniGameFakeMember()
	m.StartInsertMigrations()
	logger.Debug("------------ DB migrations Status [Done] ------------")
}

func (m *Migrations) handleTableCreation(tableName string, table interface{}) {
	err := DB.AutoMigrate(table)

	if err != nil {
		logger.Error("Error on migration", err.Error())
	}
}

func (m *Migrations) handleIndexCreation(table interface{}) {
	for _, v := range m.indexes {
		isIndexExist := DB.Migrator().HasIndex(table, v)
		if !isIndexExist {
			err := DB.Migrator().CreateIndex(table, v)
			if err != nil {
				logger.Debug("--- create--index", err)
			}
		}
	}
}
