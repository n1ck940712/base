package routefunc

import (
	"fmt"
	"net/http"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/fakemember"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"github.com/gin-gonic/gin"
)

func GetTableFakeMember(ctx *gin.Context) {
	var (
		tableID      int64
		fakeMemberID int64
	)

	fmt.Sscan(ctx.Param("table_id"), &tableID)
	fmt.Sscan(ctx.Param("fake_id"), &fakeMemberID)

	var fakeMember fakemember.IFakeMember = fakemember.NewFakeMember(tableID)

	res := fakeMember.GetTableFakeMember(fakeMemberID)

	if res == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Fake Member Not Found"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"data": res,
		})
	}
}

func DeleteTableFakeMember(ctx *gin.Context) {
	var (
		tableID      int64
		fakeMemberID int64
	)

	fmt.Sscan(ctx.Param("table_id"), &tableID)
	fmt.Sscan(ctx.Param("fake_id"), &fakeMemberID)

	var fakeMember fakemember.IFakeMember = fakemember.NewFakeMember(tableID)
	res := fakeMember.GetTableFakeMember(fakeMemberID)

	if res == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Fake member not found"})
		return
	}

	err := fakeMember.DeleteTableFakeMember(fakeMemberID)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "Fake Member Deleted"})
	}
}

func GetTableFakeMembers(ctx *gin.Context) {
	var tableID int64
	fmt.Sscan(ctx.Param("table_id"), &tableID)

	var fakeMember fakemember.IFakeMember = fakemember.NewFakeMember(tableID)
	res := fakeMember.GetTableFakeMembers(ctx)

	ctx.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}

func CreateTableFakeMember(ctx *gin.Context) {
	var tableID int64

	fmt.Sscan(ctx.Param("table_id"), &tableID)

	var fakeMember fakemember.IFakeMember = fakemember.NewFakeMember(tableID)

	var fakePayload models.FakeMember
	if err := ctx.BindJSON(&fakePayload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := fakeMember.CreateTableFakeMembers(fakePayload)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"data": res,
		})
	}
}

func UpdateTableFakeMember(ctx *gin.Context) {
	var (
		tableID      int64
		fakeMemberID int64
	)

	fmt.Sscan(ctx.Param("table_id"), &tableID)
	fmt.Sscan(ctx.Param("fake_id"), &fakeMemberID)

	var fakeMember fakemember.IFakeMember = fakemember.NewFakeMember(tableID)
	res := fakeMember.GetTableFakeMember(fakeMemberID)

	if res == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Fake member not found"})
		return
	}

	var fakePayload models.FakeMember
	if err := ctx.BindJSON(&fakePayload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res1, err := fakeMember.UpdateTableFakeMembers(fakeMemberID, fakePayload)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"data": res1,
		})
	}
}
