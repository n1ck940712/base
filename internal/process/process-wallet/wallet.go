package process_wallet

import (
	"context"
	"sync"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"github.com/google/uuid"
)

const WalletType = "wallet"
const TransactionTypeBet = "bet"

type WalletDatasouce interface {
	walletSubprocessDatasource
	GetTransactions() *[]TRequest
}

type WalletProcess interface {
	CreateTransactions() *response.ErrorData
	CommitTransactions() *response.ErrorData
	RollbackTransactions() *response.ErrorData
	CleanUp()
}

type walletProcess struct {
	datasouce    WalletDatasouce
	subprocesses walletSubprocesses
	procErr      *response.ErrorData
}

func NewWalletProcess(datasouce WalletDatasouce) WalletProcess {
	return &walletProcess{datasouce: datasouce}
}

func (wp *walletProcess) CreateTransactions() *response.ErrorData {
	if wp.procErr != nil {
		return wp.procErr
	}
	if transactions := wp.datasouce.GetTransactions(); transactions != nil && len(*transactions) > 0 {
		var wg sync.WaitGroup
		ctx, cancel := utils.CancelContext()

		for i := 0; i < len(*transactions); i++ {
			wg.Add(1)
			transaction := &(*transactions)[i]
			subprocess := newWalletSubprocess(wp.datasouce)

			wp.subprocesses = append(wp.subprocesses, subprocess)
			go func() {
				defer wg.Done()
				if err := subprocess.CreateTransaction(ctx, transaction); err != nil {
					if wp.procErr == nil { //store first error
						wp.procErr = err
					}
					cancel() //cancel all with first error
				}
			}()
		}
		wg.Wait()
	} else {
		wp.procErr = response.ErrorWithMessage(WalletType, "internal transactions is empty")
	}
	return wp.procErr
}

func (wp *walletProcess) CommitTransactions() *response.ErrorData {
	if wp.procErr != nil {
		return wp.procErr
	}
	var wg sync.WaitGroup
	ctx, cancel := utils.CancelContext()

	for i := 0; i < len(wp.subprocesses); i++ {
		subprocess := wp.subprocesses[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := subprocess.CommitTransaction(ctx); err != nil {
				if wp.procErr == nil { //store first error
					wp.procErr = err
				}
				cancel() //cancel all with first error
			}
		}()
	}
	wg.Wait()
	return wp.procErr
}

func (wp *walletProcess) RollbackTransactions() *response.ErrorData {
	if wp.procErr != nil {
		return wp.procErr
	}
	var wg sync.WaitGroup
	ctx, cancel := utils.CancelContext()

	for i := 0; i < len(wp.subprocesses); i++ {
		subprocess := wp.subprocesses[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := subprocess.RollbackTransaction(ctx); err != nil {
				if wp.procErr == nil { //store first error
					wp.procErr = err
				}
				cancel() //cancel all with first error
			}
		}()
	}
	wg.Wait()
	return wp.procErr
}

func (wsp *walletProcess) CleanUp() {
	wsp.procErr = nil
	wsp.subprocesses.CleanUp()
}

// subprocess
type walletSubprocesses []iwalletSubprocess

func (wsp *walletSubprocesses) CleanUp() {
	for i := 0; i < len(*wsp); i++ {
		(*wsp)[i].CleanUp()
	}
	wsp = nil
}

type iwalletSubprocess interface {
	CreateTransaction(ctx context.Context, transation *TRequest) *response.ErrorData
	CommitTransaction(ctx context.Context) *response.ErrorData
	RollbackTransaction(ctx context.Context) *response.ErrorData
	CleanUp()
}

type walletSubprocessDatasource interface {
	GetIdentifier() string
	GetUser() *models.User
}

type walletSubprocess struct {
	datasouce      walletSubprocessDatasource
	createResponse *TResponse
}

func newWalletSubprocess(datasouce walletSubprocessDatasource) iwalletSubprocess {
	return &walletSubprocess{datasouce: datasouce}
}

func (wsp *walletSubprocess) CreateTransaction(ctx context.Context, transation *TRequest) *response.ErrorData {
	user := wsp.datasouce.GetUser()

	if user == nil {
		return response.ErrorWithMessage(WalletType, "internal GetUser must not be nil")
	}
	if transation == nil {
		return response.ErrorWithMessage(WalletType, "internal GetTicketDetails must not be nil")
	}
	if err := api.NewAPI(settings.GetEBOAPI().String() + "/v4/wallet/").
		SetIdentifier(wsp.GetIdentifier() + " CreateTransaction").
		SetContext(ctx).
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		AddBody(transation).
		Post(&wsp.createResponse); err != nil {
		logger.Error("CreateTransaction error: ", err.Error())
		logger.Error("CreateTransaction error body: ", err.GetResponseBody())
		return response.ErrorIE(errors.WALLET_ERROR, errors.IEID_WALLET_ERROR, "bet")
	}
	return nil
}

func (wsp *walletSubprocess) CommitTransaction(ctx context.Context) *response.ErrorData {
	if wsp.createResponse == nil { //ignore where no response
		return nil
	}
	if err := api.NewAPI(settings.GetEBOAPI().String() + "/v4/wallet/" + wsp.createResponse.ID + "/commit/").
		SetIdentifier(wsp.GetIdentifier() + " CommitTransaction").
		SetContext(ctx).
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		Post(nil); err != nil {
		logger.Error("CommitTransaction error: ", err.Error())
		return response.ErrorIE(errors.WALLET_ERROR, errors.IEID_WALLET_ERROR_COMMIT, "bet")
	}
	return nil
}

func (wsp *walletSubprocess) RollbackTransaction(ctx context.Context) *response.ErrorData {
	if wsp.createResponse == nil { //ignore where no response
		return nil
	}
	wtRequest := TRequest{
		MemberID: wsp.datasouce.GetUser().EsportsID,
		RefNo:    uuid.NewString(),
	}

	if err := api.NewAPI(settings.GetEBOAPI().String() + "/v4/wallet/" + wsp.createResponse.ID + "/rollback/").
		SetIdentifier(wsp.GetIdentifier() + " RollbackTransaction").
		SetContext(ctx).
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		AddBody(wtRequest).
		Post(nil); err != nil {
		logger.Error("RollbackTransaction error: ", err.Error())
		return response.ErrorIE(errors.WALLET_ERROR, errors.IEID_WALLET_ERROR_ROLLBACK, "bet")
	}

	return nil
}

func (wsp *walletSubprocess) GetIdentifier() string {
	return wsp.datasouce.GetIdentifier()
}

func (wsp *walletSubprocess) CleanUp() {
	wsp.createResponse = nil
}
