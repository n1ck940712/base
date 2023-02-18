package api

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
)

type WalletAPI struct {
	api API
}

func NewWalletAPI() *WalletAPI {
	api := NewAPI(settings.EBO_API + "/v4/wallet")
	api.AddHeaders(map[string]string{
		"User-Agent":    settings.GetUserAgent().String(),
		"Authorization": settings.GetServerToken().String(),
		"Content-Type":  "application/json",
	})
	return &WalletAPI{api}
}

func (wapi *WalletAPI) Commit(transactionID string, response any) error {
	return wapi.api.AddPath(transactionID + "/commit/").SetIdentifier("WalletAPI Commit").Post(&response)
}

func (wapi *WalletAPI) GetTicket(ticketID string, response any) error {
	return wapi.api.SetIdentifier("WalletAPI GetTicket").AddQueries(map[string]string{"ticket_id": ticketID}).Get(&response)
}
