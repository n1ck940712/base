package settings

import (
	"os"
	"testing"
)

func TestSettingENVImageTag(t *testing.T) {
	os.Setenv("IMAGE_TAG", "minigame-backend-golang-websocket:dev-1")
	println("image tag: ", GetImageTag().String())
	println("build version: ", GetBuildVersion().String())
	os.Setenv("IMAGE_TAG", "minigame-backend-golang-websocket:dev-21")
	println("image tag: ", GetImageTag().String())
	println("build version: ", GetBuildVersion().String())
}
