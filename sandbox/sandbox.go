package sandbox

import (
	"example/web-service-gin/proto"
	"example/web-service-gin/providers"

	googleProto "google.golang.org/protobuf/proto"
)

func SendSubtractMoneyRequest() {
	request := &proto.SubtractMoneyRequest{
		UserId:   "18e79bd5-8c43-11ee-b828-080027344301",
		Currency: proto.Currency_CURRENCY_USD,
		Volume:   1,
	}
	sendBody, _ := googleProto.Marshal(request)

	rabbitProvider := providers.NewRabbitProvider()
	rabbitProvider.SendMessage("ex.BalanceService", "rk.SubtractMoneyRequest", sendBody)
}
