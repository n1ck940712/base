package errors

type finalErrorMessage struct {
	messageType string
	ied         GenericError
	ieid        IEIDError
	message     string
	mid         string
}

type FinalErrorMessage interface {
	Ied() GenericError
	Ieid() IEIDError
	MessageType() string
	Message() string
	Mid() string
}

//Getter
func (fErrorMessage *finalErrorMessage) Ied() GenericError {
	return fErrorMessage.ied
}
func (fErrorMessage *finalErrorMessage) Ieid() IEIDError {
	return fErrorMessage.ieid
}
func (fErrorMessage *finalErrorMessage) MessageType() string {
	return fErrorMessage.messageType
}
func (fErrorMessage *finalErrorMessage) Message() string {
	return fErrorMessage.message
}
func (fErrorMessage *finalErrorMessage) Mid() string {
	return fErrorMessage.mid
}

//Setter
func (fErrorMessage *finalErrorMessage) SetIed(value GenericError) {
	fErrorMessage.ied = value
}
func (fErrorMessage *finalErrorMessage) SetIeid(value IEIDError) {
	fErrorMessage.ieid = value
}
func (fErrorMessage *finalErrorMessage) SetMessageType(value string) {
	fErrorMessage.messageType = value
}
func (fErrorMessage *finalErrorMessage) SetMessage(value string) {
	fErrorMessage.message = value
}
func (fErrorMessage *finalErrorMessage) SetMid(value string) {
	fErrorMessage.mid = value
}

//constructor
func newErrorMessage(ied GenericError, ieid IEIDError, message string) FinalErrorMessage {
	return &finalErrorMessage{
		ied:         ied,
		ieid:        ieid,
		message:     message,
		messageType: "",
		mid:         "",
	}
}

//composer
func FinalizeErrorMessage(ied GenericError, ieid IEIDError, mErrMsg bool) FinalErrorMessage {
	var msg string
	if mErrMsg {
		msg = ERROR_MAP[ied]
	} else {
		msg = ERROR_MAP_IEID[ieid]
	}
	return newErrorMessage(ied, ieid, msg)
}
