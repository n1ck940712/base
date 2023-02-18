package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

var (
	db = initDB()
	DB = new(db).DB
)

type handler struct {
	DB *gorm.DB
}

func initDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(settings.DB_HOST), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		panic("Fail to open DB: " + err.Error())
	}

	useReplicaIfNeeded(db)
	return db
}

func new(db *gorm.DB) handler {
	return handler{db}
}

func useReplicaIfNeeded(db *gorm.DB) {
	if (types.Array[string]{"dev", "live"}).Constains(settings.ENVIRONMENT) && settings.DB_REPLICA != "" {
		db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(settings.DB_REPLICA)},
			Policy:   dbresolver.RandomPolicy{},
		}))
	}
}

//get first where model and update model
func Delete[M models.Model](model *M) error {
	return DB.Delete(&model).Error
}

//get first where model and update model
func Get[M models.Model](model *M) error {
	return DB.Where(model).First(&model).Error
}

//get list where model and update models
func List[M models.Model](models *[]M, whereModel *M, pagination *models.Pagination) error {
	if pagination != nil {
		offset := (pagination.Page - 1) * pagination.Limit
		queryBuider := DB.Limit(pagination.Limit).Offset(offset).Order(pagination.Sort)

		return queryBuider.Where(whereModel).Find(&models).Error
	}

	return DB.Where(whereModel).Find(models).Error
}

// create
func Create[M models.Model](model *M) error {
	return DB.Create(&model).Error
}

//update where model primary key and update model
func Update[M models.Model](model *M) error {
	return DB.Updates(model).Error
}

//update where model and update model
func UpdateWhere[M models.Model](model *M, whereModel *M) error {
	return DB.Where(whereModel).Updates(model).Error
}

func GeneratePaginationFromRequest(c *gin.Context) models.Pagination {
	// Initializing default
	limit := constants.PAGINATION_DEFAULT_LIMIT_PER_PAGE
	page := constants.PAGINATION_DEFAULT_PAGE
	sort := constants.PAGINATION_SORT_FEILD
	query := c.Request.URL.Query()

	for key, value := range query {
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			limit, _ = strconv.Atoi(queryValue)
		case "page":
			page, _ = strconv.Atoi(queryValue)
		case "sort":
			sort = queryValue

		}
	}

	return models.Pagination{
		Limit: limit,
		Page:  page,
		Sort:  sort,
	}
}
