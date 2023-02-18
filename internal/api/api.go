package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type APIError interface {
	error
	GetRequest() *http.Request
	GetResponse() *http.Response
	GetResponseBody() string
}

type apiError struct {
	err          string
	request      *http.Request
	response     *http.Response
	responseBody *[]byte
}

func newAPIError(err string, request *http.Request, response *http.Response, responseBody *[]byte) APIError {
	return &apiError{err: err, request: request, response: response, responseBody: responseBody}
}

func (ae *apiError) Error() string {
	return ae.err
}

func (ae *apiError) GetRequest() *http.Request {
	return ae.request
}

func (ae *apiError) GetResponse() *http.Response {
	return ae.response
}

func (ae *apiError) GetResponseBody() string {
	if ae.responseBody != nil {
		return string(*ae.responseBody)
	}
	return ""
}

type API interface {
	SetIdentifier(identifier string) API
	SetContext(ctx context.Context) API
	AddPath(path string) API
	AddHeaders(headers map[string]string) API
	AddQueries(queries map[string]string) API
	AddBody(body any) API
	AddTimeout(timeout time.Duration) API
	Get(response any) APIError
	GetIgnoreResponse() APIError
	Post(response any) APIError
	Patch(response any) APIError
	Put(response any) APIError
	Delete(response any) APIError
	Cancel()
}

type api struct {
	identifier string
	baseURL    string
	path       string
	headers    map[string]string
	queries    map[string]string
	body       any
	bodyStr    string
	timeout    time.Duration
	cancelCtx  context.Context
	cancelFunc *context.CancelFunc
}

func NewAPI(baseURL string) API {
	return &api{
		identifier: generateIdentifier(),
		baseURL:    baseURL,
	}
}

func (a *api) SetIdentifier(identifier string) API {
	a.identifier = identifier
	return a
}

func (a *api) SetContext(ctx context.Context) API {
	a.cancelCtx = ctx
	return a
}

func (a *api) AddPath(path string) API {
	a.path = path
	return a
}

func (a *api) AddHeaders(headers map[string]string) API {
	a.headers = headers
	return a
}

func (a *api) AddQueries(queries map[string]string) API {
	a.queries = queries
	return a
}

func (a *api) AddBody(body any) API {
	if bodyStr, ok := body.(string); ok {
		a.bodyStr = bodyStr
	} else {
		a.body = body
	}
	return a
}

func (a *api) AddTimeout(timeout time.Duration) API {
	a.timeout = timeout
	return a
}

func (a *api) Get(response any) APIError {
	return a.execute(http.MethodGet, &response)
}

func (a *api) GetIgnoreResponse() APIError {
	return a.executeIgnoreResponse(http.MethodGet)
}

func (a *api) Post(response any) APIError {
	return a.execute(http.MethodPost, &response)
}

func (a *api) Patch(response any) APIError {
	return a.execute(http.MethodPatch, &response)
}

func (a *api) Put(response any) APIError {
	return a.execute(http.MethodPut, &response)
}

func (a *api) Delete(response any) APIError {
	return a.execute(http.MethodDelete, &response)
}

func (a *api) Cancel() {
	if a.cancelFunc != nil {
		(*a.cancelFunc)()
	}
}

func (a *api) execute(method string, response any) APIError {
	ls := logger.NewLoggerStack()
	defer func() {
		ls.Print(a.identifier, " ")
		a.cleanup()
	}()
	if a.dummyExecute(&response) {
		ls.Debug(a.identifier, " loaded from dummy: ", utils.JSON(response))
		return nil
	}
	ls.Info(a.identifier, " ", method, ": ", a.getURL())
	client := &http.Client{
		Timeout: a.timeout,
	}
	jsonBody := []byte{}

	//request body
	if a.body != nil {
		body, _ := json.Marshal(a.body)
		jsonBody = body
		ls.Debug(a.identifier, " request body: ", string(jsonBody))
	} else if a.bodyStr != "" {
		jsonBody = []byte(a.bodyStr)
		ls.Debug(a.identifier, " request body: ", string(jsonBody))
	}
	ctx := a.cancelCtx

	if ctx == nil {
		internalCtx, cancelFunc := context.WithCancel(context.Background())

		ctx = internalCtx
		a.cancelFunc = &cancelFunc
	}

	req, rErr := http.NewRequestWithContext(ctx, method, a.getURL(), bytes.NewBuffer(jsonBody))

	if rErr != nil {
		ls.Error(a.identifier, " request error: ", rErr.Error())
		return newAPIError(rErr.Error(), req, nil, nil)
	}

	//headers
	if a.headers != nil {
		for k, v := range a.headers {
			req.Header.Set(k, v)
		}
	}

	//queries
	if a.queries != nil {
		q := req.URL.Query()

		for k, v := range a.queries {
			q.Add(k, v)
		}

		req.URL.RawQuery = q.Encode()
	}
	resp, cErr := client.Do(req)

	if cErr != nil {
		ls.Error(a.identifier, " request error: ", cErr.Error())
		return newAPIError(cErr.Error(), req, resp, nil)
	}
	defer resp.Body.Close()
	respBody, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		ls.Error(a.identifier, " read response error: ", readErr.Error())
		return newAPIError(readErr.Error(), req, resp, &respBody)
	}
	ls.Debug(a.identifier, " status code: ", resp.StatusCode)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if res, ok := response.(*any); ok { //ignore if response is nil
			if *res == nil {
				return nil
			}
		}
		ls.Debug(a.identifier, " response body: ", string(respBody))
		if err := json.Unmarshal(respBody, &response); err != nil {
			ls.Error(a.identifier, " json.Unmarshal response error: ", err.Error())
		}
		return nil
	}
	apiError := newAPIError(resp.Status, req, resp, &respBody)
	ls.Error(a.identifier, " request error: ", apiError.Error())
	return apiError
}

func (a *api) getURL() string {
	return a.baseURL + a.path
}

func (a *api) executeIgnoreResponse(method string) APIError {
	ls := logger.NewLoggerStack()
	defer func() {
		ls.Print(a.identifier, " ")
		a.cleanup()
	}()

	client := &http.Client{
		Timeout: a.timeout,
	}
	var jsonBody []byte

	//request body
	if a.body != nil {
		body, _ := json.Marshal(a.body)
		jsonBody = body
	}
	req, rErr := http.NewRequest(method, a.baseURL+a.path, bytes.NewBuffer(jsonBody))

	if rErr != nil {
		ls.Error(a.identifier, " ", "error request:", " ", rErr.Error())
		return newAPIError(rErr.Error(), req, nil, nil)
	}

	//headers
	if a.headers != nil {
		for k, v := range a.headers {
			req.Header.Set(k, v)
		}
	}

	//queries
	if a.queries != nil {
		q := req.URL.Query()

		for k, v := range a.queries {
			q.Add(k, v)
		}

		req.URL.RawQuery = q.Encode()
	}

	ls.Debug(a.identifier, " ", method, " ", req.URL.String())
	if jsonBody != nil {
		ls.Error(a.identifier, " ", "request:", " ", string(jsonBody))
	}
	resp, cErr := client.Do(req)

	if cErr != nil {
		ls.Error(a.identifier, " ", "error request:", " ", cErr.Error())
		return newAPIError(cErr.Error(), req, resp, nil)
	}
	defer resp.Body.Close()
	respBody, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		ls.Error(a.identifier, " ", "error response:", " ", readErr.Error())
		return newAPIError(readErr.Error(), req, resp, &respBody)
	}
	ls.Debug(a.identifier, " ", "status code:", " ", resp.StatusCode)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ls.Error(a.identifier, " ", "response:", " ", string(respBody))
		return nil
	}
	apiError := newAPIError(resp.Status, req, resp, &respBody)

	return apiError
}

func (a *api) cleanup() {
	a.identifier = generateIdentifier()
	a.queries = nil
	a.body = nil
	a.bodyStr = ""
}

func generateIdentifier() string {
	return "API" + string(types.Int(time.Now().Nanosecond()).String())
}
