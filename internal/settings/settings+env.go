package settings

import (
	"os"
	"regexp"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type StoredENV map[string]envValue

var (
	storedENV = StoredENV{}
)

type envValue interface {
	String() string
	Int() int
	Float() types.Float
}

type _envValue struct {
	raw types.String
}

func newEnvValue(raw string) envValue {
	return &_envValue{raw: types.String(raw)}
}
func (ev *_envValue) String() string {
	return *ev.raw.Ptr()
}
func (ev *_envValue) Int() int {
	return *ev.raw.Int().Ptr()
}
func (ev *_envValue) Float() types.Float {
	return ev.raw.Float()
}

func getEnv(key string) envValue {
	if value := os.Getenv(key); value != "" {
		return newEnvValue(value)
	}
	panic(key + " is required")
}

// static functions
func GetDBHost() envValue {
	return getEnv("DB_HOST")
}

func GetDBHostReplica() envValue {
	if value := os.Getenv("DB_HOST_REPLICA"); value != "" {
		return newEnvValue(value)
	}
	return newEnvValue("100")
}

func GetEBOAPI() envValue {
	return getEnv("EBO_API")
}

func GetMGCoreAPI() envValue {
	return getEnv("MG_CORE_API")
}

func GetServerToken() envValue {
	return getEnv("SERVER_TOKEN")
}

func GetMaxHashSequenceCount() envValue {
	if value := os.Getenv("DEF_MAX_SEQUENCE"); value != "" {
		return newEnvValue(value)
	}
	return newEnvValue("100")
}

func GetRedisHost() envValue {
	return getEnv("REDIS_HOST")
}

func GetUserAgent() envValue {
	return getEnv("USER_AGENT")
}

func GetSecretKey() envValue {
	if value := os.Getenv("APP_SECRET_KEY"); value != "" {
		return newEnvValue(value)
	}
	return newEnvValue("88944584a37a0a65d4716d10d3a63818da86e42092203eb5a2e9b3c667017357")
}

func GetEnvironment() envValue {
	return getEnv("ENVIRONMENT") //valid values (dev, live, local)
}

func GetLoggerLevel() envValue {
	return getEnv("LOGGER_LEVEL")
}

func GetImageTag() envValue {
	return getEnv("IMAGE_TAG")
}

func GetBuildVersion() envValue {
	imageTag := GetImageTag()

	if buildVersion, ok := storedENV[imageTag.String()]; ok {
		return buildVersion
	}
	r, _ := regexp.Compile(`\d{1,5}`)
	envValue := newEnvValue(r.FindString(imageTag.String()))
	storedENV[imageTag.String()] = envValue
	return envValue
}

func GetMGWSBaseAPIURL() envValue {
	return getEnv("MGWS_BASE_API_URL")
}
