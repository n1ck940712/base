package settings

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const REDIS_CACHE_TIMEOUT = 2
const REDIS_DEFAULT_TIMEOUT int = 120
const LOL_TOWER_CHANNEL = "loltower"
const DEFAULT_TABLE_ID = 11
const LOL_GAME_ID = 33

var LOL_CHIPSET = []float64{10, 100, 500, 1000, 5000}
var LOL_LEVELS = map[int]float64{ // euro odds
	1:  1.61,
	2:  2.68,
	3:  4.48,
	4:  7.41,
	5:  12.21,
	6:  20.42,
	7:  33.88,
	8:  56.65,
	9:  95.43,
	10: 160.4,
}

var DB_HOST = GetEnv("DB_HOST", "postgres://postgres:123456@db:5432/mini_game")
var DB_REPLICA = GetEnv("DB_HOST_REPLICA", "")
var PORT = GetEnv("PORT", "80")
var EBO_API = GetEnv("EBO_API", "https://r4pid-test-espine.r4espt.com/api")
var SERVER_TOKEN = GetEnv("SERVER_TOKEN", "0002a1d97bb7103c90b10d71eea77b2aabbddd04662d6ee166589e810247f128265a6e55b4be6843e77369ef12a1d36ab9c305ad79a09a6c9273c02595485ff199de1d6754514323f84187e0bfddea921217599268aadf1697ca7effe5536c41c4d6e251aa0cfdcfd5a4493b85679e3cde53f45fa41a00413934709e2a9bffe52f00b00ef74ac33d493a5ef96f01d38d27f370a1cb59434526ac2ad0ed88c371f3df6f3ec39c60e6be592b687101ad30aa0ad56d5cafe9cf8a0cbf1bc183cad0d40fcaa9d29c50bde7971d04100d18270194e6d96bcb80ed9d093a553811a073b18948789030265a4f0755e1eb464f2d5b54b6c38e2b6563d140843762ee11c94d571e08ab98f4e7e6ee219a520bf23ac9f2db6cb5363651e8cc915ec45809b93237")
var FUTURE_EVENTS_COUNT, _ = strconv.Atoi(GetEnv("FUTURE_EVENTS_COUNT", "3"))
var DEF_HASH_SEQUENCE_COUNT, _ = strconv.Atoi(GetEnv("DEF_HASH_SEQUENCE_COUNT", "100"))
var DEF_MAX_SEQUENCE, _ = strconv.Atoi(GetEnv("DEF_MAX_SEQUENCE", "100"))
var LOL_TOWERS_GAME_DURATION, _ = strconv.Atoi(GetEnv("LOL_TOWERS_GAME_DURATION", "14"))
var REDIS_HOST = GetEnv("REDIS_HOST", "redis:6379")
var REDIS_PASSWORD = GetEnv("REDIS_PASSWORD", "mypassword")
var WS_PORT = GetEnv("WS_PORT", "1323")
var WS_CONNECTION_EXPIRATION, _ = strconv.Atoi(GetEnv("WS_CONNECTION_EXPIRATION", "60"))
var USER_AGENT = GetEnv("USER_AGENT", "mini-game-go/20220328")
var SECRET_KEY = GetEnv("SECRET_KEY", "m1nig@me-g0l@ng")
var APP_SECRET_KEY = GetEnv("APP_SECRET_KEY", "88944584a37a0a65d4716d10d3a63818da86e42092203eb5a2e9b3c667017357")
var ENVIRONMENT = GetEnv("ENVIRONMENT", "local") //valid values (dev, live, local)
var LOGGER_LEVEL = GetEnv("LOGGER_LEVEL", "0")
var MG_CORE_API = GetEnv("MG_CORE_API", "https://r4pid-test-api.r4espt.com/api")

type Settings struct{}

type Constants struct {
	RedisCacheTimeout   int
	RedisDefaultTimeout int
	LOLTowerChannel     string
	DefaultTableID      int64
	EboAPI              string
	UserAgent           string
	ServerToken         string
	RedisHost           string
	RedisPass           string
	LOLChipset          []float64
	LOL_LEVELS          map[int]float64
	LOL_GAME_ID         int
	APP_SECRET_KEY      string
	ENVIRONMENT         string
}

func GetTimeout(key string) time.Duration {
	var i int

	switch key {
	case "cache":
		i = REDIS_CACHE_TIMEOUT
	case "timeout":
		i = REDIS_DEFAULT_TIMEOUT
	}
	return time.Duration(i) * time.Second
}

type ISettings interface {
	GetEnv(string) string
	GetTimeout(string, string) time.Duration
	Get() *Constants
}

func NewSettings() *Settings {
	return &Settings{}
}

func (s *Settings) GetEnv(key string) string {
	_ = godotenv.Load(".env")
	value := os.Getenv(key)
	if len(value) == 0 {
		return ""
	}
	return value
}

func (s *Settings) GetTimeout(key, timeUnit string) time.Duration {
	var curTime time.Duration
	var i int
	switch key {
	case "cache":
		i = REDIS_CACHE_TIMEOUT
	case "timeout":
		i = REDIS_DEFAULT_TIMEOUT
	}

	switch timeUnit {
	case "second":
		curTime = time.Duration(i) * time.Second
	default:
		curTime = time.Duration(i) * time.Second
	}

	return curTime
}

func (s *Settings) Get() *Constants {
	return &Constants{
		REDIS_CACHE_TIMEOUT,
		REDIS_DEFAULT_TIMEOUT,
		LOL_TOWER_CHANNEL,
		DEFAULT_TABLE_ID,
		EBO_API,
		USER_AGENT,
		SERVER_TOKEN,
		REDIS_HOST,
		REDIS_PASSWORD,
		LOL_CHIPSET,
		LOL_LEVELS,
		LOL_GAME_ID,
		APP_SECRET_KEY,
		ENVIRONMENT,
	}
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
