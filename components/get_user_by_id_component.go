package components

import (
	"example/web-service-gin/proto"

	logger "github.com/sirupsen/logrus"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	getUserByIdResponsetRkName = "rk.GetUserByIdResponse"
	bsGotGetUserByIdRequestMsg = "BalanceService got GetUserByIdRequest with id: %s"
)

func (b *BalanceComponent) GetUserById(byteRequest []byte) {
	userId := string(byteRequest)
	logger.Infof(bsGotGetUserByIdRequestMsg, userId)

	user, err := b.PgProvider.GetUserById(userId)
	getUserByIdResponse := proto.GetUserByIdResponse{User: user}
	if err != nil {
		getUserByIdResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		logger.Errorf(publishedResponseErrMsg, getUserByIdResponse.String())
	} else {
		logger.Infof(publishedResponseMsg, getUserByIdResponse.String())
	}

	sendBody, _ := googleProto.Marshal(&getUserByIdResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, getUserByIdResponsetRkName, sendBody)
}
