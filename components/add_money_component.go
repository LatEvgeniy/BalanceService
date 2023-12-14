package components

import (
	"example/web-service-gin/proto"
	"example/web-service-gin/providers"
	"example/web-service-gin/validators"

	logger "github.com/sirupsen/logrus"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	balanceServiceExchangeName = "ex.BalanceService"

	publishedResponseErrMsg = "BalanceService published error msg: %+v"
	publishedResponseMsg    = "BalanceService published msg: %+v"

	addMoneyResponseRkName  = "rk.AddMoneyResponse"
	bsGotAddMoneyRequestMsg = "BalanceService got AddMoneyRequest: %+v"
)

type BalanceComponent struct {
	RabbitProvider *providers.RabbitProvider
	PgProvider     *providers.PostgresProvider
}

func (b *BalanceComponent) AddMoney(byteRequest []byte) {
	var addMoneyRequest proto.AddMoneyRequest
	if err := googleProto.Unmarshal(byteRequest, &addMoneyRequest); err != nil {
		b.sendAddMoneyErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INTERNAL, Message: err.Error()})
		return
	}
	logger.Infof(bsGotAddMoneyRequestMsg, addMoneyRequest.String())

	if err := validators.ValidateAddMoneyRequest(&addMoneyRequest); err != nil {
		b.sendAddMoneyErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INVALID_REQUEST, Message: err.Error()})
		return
	}

	if err := b.PgProvider.AddMoney(&addMoneyRequest); err != nil {
		b.sendAddMoneyErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()})
		return
	}

	user, err := b.PgProvider.GetUserById(addMoneyRequest.UserId)
	addMoneyResponse := &proto.AddMoneyResponse{User: user}
	if err != nil {
		addMoneyResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		logger.Errorf(publishedResponseErrMsg, addMoneyResponse)
	} else {
		logger.Infof(publishedResponseMsg, addMoneyResponse)
	}

	sendBody, _ := googleProto.Marshal(addMoneyResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, addMoneyResponseRkName, sendBody)
}

func (b *BalanceComponent) sendAddMoneyErrResponse(err *proto.ErrorDto) {
	addMoneyErrResponse := &proto.AddMoneyResponse{Error: err}
	sendBody, _ := googleProto.Marshal(addMoneyErrResponse)
	logger.Errorf(publishedResponseErrMsg, addMoneyErrResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, addMoneyResponseRkName, sendBody)
}
