package main

import (
	"example/web-service-gin/components"
	provider "example/web-service-gin/providers"
)

var (
	balanceServiceExchangeName = "ex.BalanceService"

	addMoneyRequestRkName      = "rk.AddMoneyRequest"
	subtractMoneyRequestRkName = "rk.SubtractMoneyRequest"
	createUserRequestRkName    = "rk.CreateUserRequest"
	getUserRequestRkName       = "rk.GetUserByIdRequest"
	lockBalanceRequestRkName   = "rk.LockBalanceRequest"
	unlockBalanceRequestRkName = "rk.UnlockBalanceRequest"

	addMoneyQueueName      = "q.BalanceService.AddMoneyRequest.Listener"
	subtractMoneyQueueName = "q.BalanceService.SubtractMoneyRequest.Listener"
	createUserQueueName    = "q.BalanceService.CreateUserRequest.Listener"
	getUserQueueName       = "q.BalanceService.GetUserByIdRequest.Listener"
	lockBalanceQueueName   = "q.BalanceService.LockBalanceRequest.Listener"
	unlockBalanceQueueName = "q.BalanceService.UnlockBalanceRequest.Listener"
)

func main() {
	rabbitProvider := provider.NewRabbitProvider()
	pgProvider := provider.PostgresProvider{}

	BalanceComponent := &components.BalanceComponent{RabbitProvider: rabbitProvider, PgProvider: &pgProvider}

	rabbitProvider.DeclareExchange(balanceServiceExchangeName)

	runListeners(rabbitProvider, BalanceComponent)
}

func runListeners(rabbitProvider *provider.RabbitProvider, BalanceComponent *components.BalanceComponent) {
	msgs, ch := rabbitProvider.GetQueueConsumer(balanceServiceExchangeName, addMoneyRequestRkName, addMoneyQueueName)
	go rabbitProvider.RunListener(msgs, ch, BalanceComponent.AddMoney)

	msgs, ch = rabbitProvider.GetQueueConsumer(balanceServiceExchangeName, subtractMoneyRequestRkName, subtractMoneyQueueName)
	go rabbitProvider.RunListener(msgs, ch, BalanceComponent.SubtractMoney)

	msgs, ch = rabbitProvider.GetQueueConsumer(balanceServiceExchangeName, createUserRequestRkName, createUserQueueName)
	go rabbitProvider.RunListener(msgs, ch, BalanceComponent.CreateUser)

	msgs, ch = rabbitProvider.GetQueueConsumer(balanceServiceExchangeName, getUserRequestRkName, getUserQueueName)
	go rabbitProvider.RunListener(msgs, ch, BalanceComponent.GetUserById)

	msgs, ch = rabbitProvider.GetQueueConsumer(balanceServiceExchangeName, lockBalanceRequestRkName, lockBalanceQueueName)
	go rabbitProvider.RunListener(msgs, ch, BalanceComponent.LockBalance)

	msgs, ch = rabbitProvider.GetQueueConsumer(balanceServiceExchangeName, unlockBalanceRequestRkName, unlockBalanceQueueName)
	rabbitProvider.RunListener(msgs, ch, BalanceComponent.UnlockBalance)
}
