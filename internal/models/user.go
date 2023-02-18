package models

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type ESUser struct {
	ID       int64      `json:"id"` //esports_id
	Type     string     `json:"type"`
	Metadata ESMetadata `json:"metadata"`
}

type ESMetadata struct {
	MemberCode      string  `json:"member_code"`
	PartnerID       int32   `json:"partner_id"`
	IsAccountFrozen bool    `json:"is_account_frozen"`
	CurrencyCode    string  `json:"currency_code"`
	ExchangeRate    float32 `json:"exchange_rate"`
	CurrencyRation  float32 `json:"currency_ration"`
}

type User struct {
	ID               int64        `gorm:"column:id;primaryKey" json:"id"`
	Password         string       `gorm:"column:password;->:false;<-:create" json:"password"`
	IsSuperUser      bool         `gorm:"column:is_superuser;default:false" json:"is_superuser"`
	Username         string       `gorm:"column:username" json:"username"`
	FirstName        string       `gorm:"column:first_name;default:''" json:"first_name"`
	LastName         string       `gorm:"column:last_name;default:''" json:"last_name"`
	Email            string       `gorm:"column:email" json:"email"`
	IsStaff          bool         `gorm:"column:is_staff;default:false" json:"is_staff"`
	IsActive         bool         `gorm:"column:is_active;default:true" json:"is_active"`
	DateJoined       time.Time    `gorm:"column:date_joined;autoCreateTime" json:"date_joined"`
	UserType         int8         `gorm:"column:user_type" json:"user_type"`
	EsportsID        int64        `gorm:"column:esports_id" json:"esports_id"`
	EsportsPartnerID int64        `gorm:"column:esports_partner_id" json:"esports_partner_id"`
	IsAccountFrozen  bool         `gorm:"column:is_account_frozen;default:false" json:"is_account_frozen"`
	MemberCode       string       `gorm:"column:member_code" json:"member_code"`
	CurrencyCode     string       `gorm:"column:currency_code" json:"currency_code"`
	ExchangeRate     float64      `gorm:"column:exchange_rate" json:"exchange_rate"`
	SleepStatus      int8         `gorm:"column:sleep_status" json:"sleep_status"`
	CurrencyRatio    float64      `gorm:"column:currency_ratio" json:"currency_ratio"`
	UserRequest      *UserRequest `gorm:"-" json:"-"`
	AuthToken        *string      `gorm:"-" json:"-"`
}

func (u *User) GetRequestIPAddress() string {
	if u.UserRequest != nil {
		return u.UserRequest.IpAddress
	}
	return "0.0.0.0"
}

func (u *User) GetRequestSource() *string {
	if u.UserRequest != nil {
		return &u.UserRequest.Source
	}
	unknown := "unknown"

	return &unknown
}

func (u *User) GetRequestUserAgent() *string {
	if u.UserRequest != nil {
		return &u.UserRequest.UserAgent
	}
	return nil
}

func (u *User) SetRequest(request *http.Request) {
	u.UserRequest = NewUserRequest(request)
}

func (User) TableName() string {
	return "mini_game_user"
}

type UserRequest struct {
	IpAddress string
	Source    string
	UserAgent string
	Request   *http.Request
}

func NewUserRequest(request *http.Request) *UserRequest {
	userRequest := UserRequest{
		IpAddress: "0.0.0.0",
		Source:    "unknown",
	}

	userRequest.setIPAddress(request)
	userRequest.setRequestSource(request)
	return &userRequest
}

func (ur *UserRequest) setIPAddress(r *http.Request) {
	if len(r.Header) > 0 {
		if header := r.Header.Get("X-Forwarded-For"); header != "" {
			// x-forwarded-for may return multiple values in the format
			// @see https://en.wikipedia.org/wiki/X-Forwarded-For#Format
			proxies := strings.Split(header, ",")

			for _, proxy := range proxies {
				proxyIP, _ := ur.removePort(proxy)

				if ur.isIPAddress(proxyIP) {
					ur.IpAddress = proxyIP
					return
				}
			}
		}

	}
	remoteIP, err := ur.removePort(r.RemoteAddr)

	if err != nil {
		logger.Info("UserRequest removePort error: ", err.Error())
		return
	}
	if ur.isIPAddress(remoteIP) {
		ur.IpAddress = remoteIP
	}
}

func (ur *UserRequest) setRequestSource(r *http.Request) {
	const (
		desktopBrowser = "desktop-browser"
		mobileApp      = "mobile-app"
		mobileBrowser  = "mobile-browser"
	)
	type userAgents struct {
		tag    string
		regexp string
	}
	ur.UserAgent = r.UserAgent()
	checklist := []userAgents{
		{desktopBrowser, `(Macintosh|Windows)`},
		{desktopBrowser, `(Linux)`},
		{mobileBrowser, `(Linux.*?Android)`},
		{mobileBrowser, `(iPhone|iPad|Linux.*?Android|Android.*?Mobile)`},
		{mobileApp, `(app-Android|app-iOS)`},
		{mobileApp, `?r4pid-esports-core\/(?P<app_name>.*?)\/1.0.0\/(?P<model>.*?)\/(?P<os_version>.*?)$`},
	}
	contains := types.Array[string]{"Android", "iPhone", "iOS", "iPad", "Phone"}.Constains(ur.UserAgent)

	for _, agent := range checklist {
		isMatch, _ := regexp.MatchString("(?i)"+agent.regexp, ur.UserAgent)

		if contains && isMatch && (agent.tag == mobileBrowser || agent.tag == mobileApp) {
			ur.Source = agent.tag
			break
		} else if !contains && isMatch {
			ur.Source = agent.tag
			break
		}
	}
}

func (ur *UserRequest) isIPAddress(address string) bool {
	return net.ParseIP(address) != nil
}

func (ur *UserRequest) removePort(address string) (string, error) {
	host, _, err := net.SplitHostPort(address)

	if err == nil { //no error or has proxy
		return host, err
	}
	return address, err //has error or has no proxy
}
