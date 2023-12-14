package components

import (
	"example/web-service-gin/proto"
	"example/web-service-gin/validators"

	logger "github.com/sirupsen/logrus"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	subtractMoneyResponseRkName  = "rk.SubtractMoneyResponse"
	bsGotSubtractMoneyRequestMsg = "BalanceService got SubtractMoneyRequest: %s"
)

func (b *BalanceComponent) SubtractMoney(byteRequest []byte) {
	var subtractMoneyRequest proto.SubtractMoneyRequest
	if err := googleProto.Unmarshal(byteRequest, &subtractMoneyRequest); err != nil {
		b.sendSubtractMoneyErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INTERNAL, Message: err.Error()})
		return
	}
	logger.Infof(bsGotSubtractMoneyRequestMsg, subtractMoneyRequest.String())

	if err := validators.ValidateSubtractMoneyRequest(&subtractMoneyRequest); err != nil {
		b.sendSubtractMoneyErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INVALID_REQUEST, Message: err.Error()})
		return
	}

	if err := b.PgProvider.SubtractMoney(&subtractMoneyRequest); err != nil {
		b.sendSubtractMoneyErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()})
		return
	}

	user, err := b.PgProvider.GetUserById(subtractMoneyRequest.UserId)
	subtractMoneyResponse := &proto.SubtractMoneyResponse{User: user}
	if err != nil {
		subtractMoneyResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		logger.Errorf(publishedResponseErrMsg, subtractMoneyResponse)
	} else {
		logger.Infof(publishedResponseMsg, subtractMoneyResponse)
	}

	sendBody, _ := googleProto.Marshal(subtractMoneyResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, subtractMoneyResponseRkName, sendBody)
}

func (b *BalanceComponent) sendSubtractMoneyErrResponse(err *proto.ErrorDto) {
	subtractMoneyErrResponse := &proto.SubtractMoneyResponse{Error: err}
	sendBody, _ := googleProto.Marshal(subtractMoneyErrResponse)
	logger.Errorf(publishedResponseErrMsg, subtractMoneyErrResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, subtractMoneyResponseRkName, sendBody)
}
