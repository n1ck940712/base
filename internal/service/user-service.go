package service

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm"
)

type UserService struct{}

type IUser interface {
	GetMGDetails(id int64) models.User
}

func NewUser() *UserService {
	return &UserService{}
}

func (u *UserService) GetMGDetails(esId int64) (models.User, error) {
	var user models.User
	result := DB.Table(`mini_game_user`).
		Where("esports_id = ?", esId).
		Where("is_account_frozen = ?", false).
		First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Debug("User Not found!")
	}

	return user, result.Error
}
