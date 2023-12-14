package validators

import (
	"errors"
	"example/web-service-gin/proto"

	"github.com/google/uuid"
)

var (
	notImplementedCurrencyErrorMsg = "not implemented request currency"
)

func ValidateAddMoneyRequest(request *proto.AddMoneyRequest) error {
	if request.Volume <= 0 {
		return errors.New("volume must be greater than 0")
	}

	_, err := uuid.Parse(request.UserId)
	if err != nil {
		return err
	}

	if err := isCurrencyImplemented(request.Currency); err != nil {
		return err
	}

	return nil
}

func ValidateSubtractMoneyRequest(request *proto.SubtractMoneyRequest) error {
	if request.Volume <= 0 {
		return errors.New("volume must be greater than 0")
	}

	_, err := uuid.Parse(request.UserId)
	if err != nil {
		return err
	}

	if err := isCurrencyImplemented(request.Currency); err != nil {
		return err
	}

	return nil
}

func ValidateLockBalanceRequest(request *proto.LockBalanceRequest) error {
	if request.Volume <= 0 {
		return errors.New("volume must be greater than 0")
	}
	_, err := uuid.Parse(request.UserId)
	if err != nil {
		return err
	}

	if err := isCurrencyImplemented(request.Currency); err != nil {
		return err
	}

	return nil
}

func ValidateUnlockBalanceRequest(request *proto.UnlockBalanceRequest) error {
	if request.Volume <= 0 {
		return errors.New("volume must be greater than 0")
	}

	_, err := uuid.Parse(request.UserId)
	if err != nil {
		return err
	}

	if err := isCurrencyImplemented(request.Currency); err != nil {
		return err
	}

	return nil
}

func isCurrencyImplemented(currency proto.Currency) error {
	switch currency {
	case proto.Currency_CURRENCY_EUR:
		return nil
	case proto.Currency_CURRENCY_UAH:
		return nil
	case proto.Currency_CURRENCY_USD:
		return nil
	default:
		return errors.New(notImplementedCurrencyErrorMsg)
	}
}
