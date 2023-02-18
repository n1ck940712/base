package response

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
)

func ErrorGE(ied errors.GenericError, ieid errors.IEIDError, eType string) *ErrorData {
	return &ErrorData{
		Ied:  int32(ied),
		Ieid: int32(ieid),
		Mid:  0,
		Msg:  errors.ERROR_MAP[ied],
		Type: eType,
	}
}

func ErrorGEMessage(ied errors.GenericError, ieid errors.IEIDError, message string, eType string) *ErrorData {
	return &ErrorData{
		Ied:  int32(ied),
		Ieid: int32(ieid),
		Mid:  0,
		Msg:  message,
		Type: eType,
	}
}

func ErrorIE(ied errors.GenericError, ieid errors.IEIDError, eType string) *ErrorData {
	return &ErrorData{
		Ied:  int32(ied),
		Ieid: int32(ieid),
		Mid:  0,
		Msg:  errors.ERROR_MAP_IEID[ieid],
		Type: eType,
	}
}

func ErrorIEMessage(ied errors.GenericError, ieid errors.IEIDError, message string, eType string) *ErrorData {
	return &ErrorData{
		Ied:  int32(ied),
		Ieid: int32(ieid),
		Mid:  0,
		Msg:  message,
		Type: eType,
	}
}

func ErrorBadRequest(eType string) *ErrorData {
	return &ErrorData{
		Ied:  0,
		Ieid: 0,
		Mid:  0,
		Msg:  "bad request",
		Type: eType,
	}
}

func ErrorInValidType(eType string) *ErrorData {
	return &ErrorData{
		Ied:  0,
		Ieid: 0,
		Mid:  0,
		Msg:  "invalid type",
		Type: eType,
	}
}

func ErrorNotImplemented(eType string) *ErrorData {
	return &ErrorData{
		Ied:  0,
		Ieid: 0,
		Mid:  0,
		Msg:  "not implemented",
		Type: eType,
	}
}

func ErrorWithMessage(message string, pType string) *ErrorData {
	return &ErrorData{
		Ied:  0,
		Ieid: 0,
		Type: pType,
		Msg:  message,
		Mid:  0,
	}
}

type ErrorData struct {
	Ied  int32  `json:"ied"`
	Ieid int32  `json:"ieid"`
	Mid  int32  `json:"mid"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

func (err *ErrorData) Description() string {
	jStr, _ := json.MarshalIndent(err, "", "    ")

	return string(jStr)
}

func (err *ErrorData) Error() string {
	return utils.JSON(err)
}
