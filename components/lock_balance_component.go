package components

import (
	"example/web-service-gin/proto"
	"example/web-service-gin/validators"

	logger "github.com/sirupsen/logrus"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	lockBalanceResponsetRkName = "rk.LockBalanceResponse"
	bsGotLockBalanceRequestMsg = "BalanceService got LockBalanceRequest: %+v"
)

func (b *BalanceComponent) LockBalance(byteRequest []byte) {
	var lockBalanceRequest proto.LockBalanceRequest

	if err := googleProto.Unmarshal(byteRequest, &lockBalanceRequest); err != nil {
		b.sendLockBalanceErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INTERNAL, Message: err.Error()})
		return
	}

	logger.Infof(bsGotLockBalanceRequestMsg, lockBalanceRequest.String())

	if err := validators.ValidateLockBalanceRequest(&lockBalanceRequest); err != nil {
		b.sendLockBalanceErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INVALID_REQUEST, Message: err.Error()})
		return
	}

	lockBalanceResponse := &proto.LockBalanceResponse{UserId: lockBalanceRequest.UserId}
	if err := b.PgProvider.LockBalance(&lockBalanceRequest); err != nil {
		lockBalanceResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		logger.Errorf(publishedResponseErrMsg, lockBalanceResponse)
	} else {
		logger.Infof(publishedResponseMsg, lockBalanceResponse)
	}

	sendBody, _ := googleProto.Marshal(lockBalanceResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, lockBalanceResponsetRkName, sendBody)
}

func (b *BalanceComponent) sendLockBalanceErrResponse(err *proto.ErrorDto) {
	lockBalanceResponse := &proto.LockBalanceResponse{Error: err}
	sendBody, _ := googleProto.Marshal(lockBalanceResponse)
	logger.Errorf(publishedResponseErrMsg, lockBalanceResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, lockBalanceResponsetRkName, sendBody)
}
