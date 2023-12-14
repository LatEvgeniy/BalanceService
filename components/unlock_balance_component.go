package components

import (
	"example/web-service-gin/proto"
	"example/web-service-gin/validators"

	logger "github.com/sirupsen/logrus"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	unlockBalanceResponsetRkName = "rk.UnlockBalanceResponse"
	bsGotUnlockBalanceRequestMsg = "BalanceService got UnlockBalanceRequest: %+v"
)

func (b *BalanceComponent) UnlockBalance(byteRequest []byte) {
	var unlockBalanceRequest proto.UnlockBalanceRequest

	if err := googleProto.Unmarshal(byteRequest, &unlockBalanceRequest); err != nil {
		b.sendUnLockBalanceErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INTERNAL, Message: err.Error()})
		return
	}

	logger.Infof(bsGotUnlockBalanceRequestMsg, unlockBalanceRequest.String())

	if err := validators.ValidateUnlockBalanceRequest(&unlockBalanceRequest); err != nil {
		b.sendUnLockBalanceErrResponse(&proto.ErrorDto{Code: proto.ErrorCode_ERROR_INVALID_REQUEST, Message: err.Error()})
		return
	}

	unlockBalanceResponse := &proto.UnlockBalanceResponse{UserId: unlockBalanceRequest.UserId}
	if err := b.PgProvider.UnlockBalance(&unlockBalanceRequest); err != nil {
		unlockBalanceResponse.Error = &proto.ErrorDto{Code: proto.ErrorCode_ERROR_POSTGRES_PROCESSING, Message: err.Error()}
		logger.Errorf(publishedResponseErrMsg, unlockBalanceResponse)
	} else {
		logger.Infof(publishedResponseMsg, unlockBalanceResponse)
	}

	sendBody, _ := googleProto.Marshal(unlockBalanceResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, unlockBalanceResponsetRkName, sendBody)
}

func (b *BalanceComponent) sendUnLockBalanceErrResponse(err *proto.ErrorDto) {
	lockBalanceResponse := &proto.LockBalanceResponse{Error: err}
	sendBody, _ := googleProto.Marshal(lockBalanceResponse)
	logger.Errorf(publishedResponseErrMsg, lockBalanceResponse)
	b.RabbitProvider.SendMessage(balanceServiceExchangeName, unlockBalanceResponsetRkName, sendBody)
}
