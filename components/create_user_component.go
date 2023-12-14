package components

import (
	"example/web-service-gin/proto"

	logger "github.com/sirupsen/logrus"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	createUserResponseRkName  = "rk.CreateUserResponse"
	bsGotCreateUserRequestMsg = "BalanceService got CreateUserRequest with name: %s"
)

func (b *BalanceComponent) CreateUser(byteRequest []byte) {
	userName := string(byteRequest)
	createUserResponse := proto.CreateUserResponse{}
	logger.Infof(bsGotCreateUserRequestMsg, userName)

	userId, err := b.PgProvider.CreateUser(userName)
	if err != nil {
		createUserResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		sendBody, _ := googleProto.Marshal(&createUserResponse)
		logger.Errorf(publishedResponseErrMsg, createUserResponse.String())
		b.RabbitProvider.SendMessage(balanceServiceExchangeName, createUserResponseRkName, sendBody)
		return
	}

	user, err := b.PgProvider.GetUserById(userId)
	createUserResponse.CreatedUser = user
	if err != nil {
		createUserResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		logger.Errorf(publishedResponseErrMsg, createUserResponse.String())
	} else {
		logger.Infof(publishedResponseMsg, createUserResponse.String())
	}

	sendBody, _ := googleProto.Marshal(&createUserResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, createUserResponseRkName, sendBody)
}
