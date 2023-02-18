package fakemember

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"github.com/gin-gonic/gin"
)

type IFakeMember interface {
	GetTableFakeMember(fakeMemberID int64) *models.FakeMember
	GetTableFakeMembers(c *gin.Context) []models.FakeMember
	CreateTableFakeMembers(fakeMem models.FakeMember) (models.FakeMember, error)
	UpdateTableFakeMembers(ffakeMemberID int64, akeMem models.FakeMember) (models.FakeMember, error)
	DeleteTableFakeMember(fakeMemberID int64) error
}
