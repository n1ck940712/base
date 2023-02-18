package fakemember

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type tableFakeMember struct {
	tableID int64
}

func NewFakeMember(tableID int64) *tableFakeMember {
	return &tableFakeMember{
		tableID: tableID,
	}
}

func (f *tableFakeMember) DeleteTableFakeMember(fakeMemberID int64) error {
	fakeMember := models.FakeMember{}
	fakeMember.ID = fakeMemberID

	if err := service.Delete(&fakeMember); err != nil {
		return err
	}

	return nil
}

func (f *tableFakeMember) GetTableFakeMember(fakeMemberID int64) *models.FakeMember {
	fakeMember := models.FakeMember{}
	fakeMember.ID = fakeMemberID

	if err := service.Get(&fakeMember); err == gorm.ErrRecordNotFound {
		return nil
	}

	return &fakeMember
}

func (f *tableFakeMember) GetTableFakeMembers(c *gin.Context) []models.FakeMember {
	fakeMembers := []models.FakeMember{}
	whereModel := models.FakeMember{
		MiniGameTableID: f.tableID,
	}
	pagination := service.GeneratePaginationFromRequest(c)

	service.List(&fakeMembers, &whereModel, &pagination)

	return fakeMembers
}

func (f *tableFakeMember) CreateTableFakeMembers(fakeMem models.FakeMember) (models.FakeMember, error) {
	fakeMem.MiniGameTableID = f.tableID
	err := service.Create(&fakeMem)

	return fakeMem, err
}

func (f *tableFakeMember) UpdateTableFakeMembers(fakeMemberID int64, fakeMem models.FakeMember) (models.FakeMember, error) {
	fakeMem.ID = fakeMemberID
	err := service.Update(&fakeMem)

	return fakeMem, err
}
