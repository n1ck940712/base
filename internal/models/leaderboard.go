package models

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type LeaderBoard struct {
	UserID     int64  `gorm:"column:user_id"`
	ID         int64  `gorm:"column:esports_id"`
	PartnetID  int64  `gorm:"column:esports_partner_id"`
	MemberCode string `gorm:"column:member_code"`
	Level      string `gorm:"column:level"`
}

func (lb *LeaderBoard) GetEncryptedUserID() string {
	return types.Bytes(fmt.Sprintf("%v%v", lb.ID, lb.PartnetID)).SHA256()
}

func (lb *LeaderBoard) GetEncryptedName() string {
	return string(types.String(lb.MemberCode).Mask("*", constants.MEMBER_CODE_MASK_COUNT, 4))
}
