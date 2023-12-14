package providers

import (
	"example/web-service-gin/proto"
	"example/web-service-gin/utils"
	"fmt"

	"context"
	"errors"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	logger "github.com/sirupsen/logrus"
)

var (
	balacneConnectionString = "postgres://username:password@postgres:5432/BalanceService"

	eurCurrencyName = "EUR"
	uahCurrencyName = "UAH"
	usdCurrencyName = "USD"
	currencyNames   = []string{uahCurrencyName, eurCurrencyName, usdCurrencyName}

	notImplementedCurrencyErrorMsg     = "not implemented request currency"
	notEnoughBalanceToLockErrorMsg     = "not enough balance to lock"
	notEnoughBalanceToUnLockErrorMsg   = "not enough balance to unlock"
	notEnoughBalanceToSubtractErrorMsg = "not enough balance to subtract"
	userNotFoundErrMsg                 = "User with id: %s not found"

	successfullyAddedMoneyMsg      = "Pg provider successfully added money for user with id %s"
	successfullySubtractedMoneyMsg = "Pg provider successfully subtracted money for user with id %s"
	successfullyCreatedUserMsg     = "Pg provider successfully created user with id %s"
	successfullyGotUserMsg         = "Pg provider successfully got user with id %s"

	insertUserQuery    = "INSERT INTO users (id, name) VALUES ($1, $2)"
	insertBalanceQuery = "INSERT INTO balances (userid, currencyid, balance, lockedbalance) VALUES ($1, $2, $3, $4)"

	addBalanceQuery      = "UPDATE balances SET balance = balance + $1 WHERE userid = $2 AND currencyid = $3;"
	subtractBalanceQuery = "UPDATE balances SET balance = balance - $1 WHERE userid = $2 AND currencyid = $3;"
	lockBalanceQuery     = "UPDATE balances SET lockedbalance = lockedbalance + $1 WHERE userid = $2 AND currencyid = $3"
	unlockBalanceQuery   = "UPDATE balances SET lockedbalance = lockedbalance - $1 WHERE userid = $2 AND currencyid = $3"

	getCurrencyIdByNameQuery = "SELECT id FROM currencies WHERE name = $1"
	getCurrencyNameByIdQuery = "SELECT name FROM currencies WHERE id = $1"
	getUsersTableQuery       = "SELECT name FROM users WHERE id = $1"
	getBalanceTableQuery     = "SELECT userid, currencyid, balance, lockedbalance FROM balances WHERE userid = $1"
	getBalancesQuery         = "SELECT balance, lockedbalance FROM balances WHERE userid = $1 AND currencyid = $2"
)

type balancesEntity struct {
	userid        string
	currencyid    string
	balance       float64
	lockedBalance float64
}

type PostgresProvider struct {
}

func (p *PostgresProvider) GetNewConnection() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), balacneConnectionString)
	utils.CheckErrorWithPanic(err)
	return conn
}

func (p *PostgresProvider) getCurrencyId(currency proto.Currency) (string, error) {
	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	currencyName, err := p.getCurrencyNameByProtoCurrency(currency)
	if err != nil {
		return "", err
	}

	var currencyId string
	if err = conn.QueryRow(context.Background(), getCurrencyIdByNameQuery, currencyName).Scan(&currencyId); err != nil {
		return "", err
	}

	return currencyId, nil
}

// --------- Add Money ---------

func (p *PostgresProvider) AddMoney(request *proto.AddMoneyRequest) error {
	currencyId, err := p.getCurrencyId(request.Currency)
	if err != nil {
		return err
	}

	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	if _, err = conn.Exec(context.Background(), addBalanceQuery, request.Volume, request.UserId, currencyId); err != nil {
		return fmt.Errorf(userNotFoundErrMsg, request.UserId)
	}

	logger.Debugf(successfullyAddedMoneyMsg, request.UserId)
	return nil
}

// --------- Subtract Money ---------

func (p *PostgresProvider) SubtractMoney(request *proto.SubtractMoneyRequest) error {
	currencyId, err := p.getCurrencyId(request.Currency)
	if err != nil {
		return err
	}

	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	var balance, lockedBalance float64
	if err = conn.QueryRow(context.Background(), getBalancesQuery, request.UserId, currencyId).Scan(&balance, &lockedBalance); err != nil {
		return fmt.Errorf(userNotFoundErrMsg, request.UserId)
	}

	if balance-lockedBalance < request.Volume {
		return errors.New(notEnoughBalanceToSubtractErrorMsg)
	}

	if _, err = conn.Exec(context.Background(), subtractBalanceQuery, request.Volume, request.UserId, currencyId); err != nil {
		return err
	}

	logger.Debugf(successfullySubtractedMoneyMsg, request.UserId)
	return nil
}

// --------- Create User ---------

func (p *PostgresProvider) CreateUser(userName string) (string, error) {
	userId, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	if err = p.insertToUserTable(userId, userName); err != nil {
		return "", err
	}

	err = p.insertToBalancesTable(userId)

	return userId.String(), err
}

func (p *PostgresProvider) insertToUserTable(userId uuid.UUID, userName string) error {
	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	conn.QueryRow(context.TODO(), insertUserQuery, userId, userName)

	return nil
}

func (p *PostgresProvider) insertToBalancesTable(userId uuid.UUID) error {
	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	for _, currencyName := range currencyNames {
		var currencyId string
		if err := conn.QueryRow(context.Background(), getCurrencyIdByNameQuery, currencyName).Scan(&currencyId); err != nil {
			return err
		}

		if _, err := conn.Exec(context.Background(), insertBalanceQuery, userId.String(), currencyId, 0, 0); err != nil {
			return err
		}
	}
	logger.Debugf(successfullyCreatedUserMsg, userId.String())
	return nil
}

// --------- Get User By Id ---------

func (p *PostgresProvider) GetUserById(userId string) (*proto.User, error) {
	var user proto.User = proto.User{Id: userId}

	if err := p.getDataFromUsersTable(&user); err != nil {
		return &user, err
	}
	if err := p.getDataFromBalancesTable(&user); err != nil {
		return &user, err
	}

	logger.Debugf(successfullyGotUserMsg, userId)
	return &user, nil
}

func (p *PostgresProvider) getDataFromUsersTable(user *proto.User) error {
	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	if err := conn.QueryRow(context.Background(), getUsersTableQuery, user.Id).Scan(&user.Name); err != nil {
		return fmt.Errorf(userNotFoundErrMsg, user.Id)
	}
	return nil
}

func (p *PostgresProvider) getDataFromBalancesTable(user *proto.User) error {
	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), getBalanceTableQuery, user.Id)
	if err != nil {
		return fmt.Errorf(userNotFoundErrMsg, user.Id)
	}
	defer rows.Close()

	var balancesEnity []balancesEntity
	for rows.Next() {
		var balance balancesEntity
		if err := rows.Scan(&balance.userid, &balance.currencyid, &balance.balance, &balance.lockedBalance); err != nil {
			return err
		}
		balancesEnity = append(balancesEnity, balance)
	}

	var balances []*proto.Balance
	for _, entity := range balancesEnity {
		protoBalance, converErr := p.convertBalancesEnityToBalancesProto(entity)
		if converErr != nil {
			return converErr
		}
		balances = append(balances, protoBalance)
	}

	user.Balances = balances
	return nil
}

func (p *PostgresProvider) convertBalancesEnityToBalancesProto(entity balancesEntity) (*proto.Balance, error) {
	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	var currencyName string
	if err := conn.QueryRow(context.Background(), getCurrencyNameByIdQuery, entity.currencyid).Scan(&currencyName); err != nil {
		return nil, err
	}
	return &proto.Balance{
		CurrencyName:  currencyName,
		Balance:       entity.balance,
		LockedBalance: entity.lockedBalance,
	}, nil
}

// --------- Lock Balance ---------

func (p *PostgresProvider) LockBalance(request *proto.LockBalanceRequest) error {
	currencyId, err := p.getCurrencyId(request.Currency)
	if err != nil {
		return err
	}

	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	var balance, lockedBalance float64
	if err = conn.QueryRow(context.Background(), getBalancesQuery, request.UserId, currencyId).Scan(&balance, &lockedBalance); err != nil {
		return fmt.Errorf(userNotFoundErrMsg, request.UserId)
	}

	if lockedBalance+request.Volume > balance {
		return errors.New(notEnoughBalanceToLockErrorMsg)
	}

	if _, err = conn.Exec(context.Background(), lockBalanceQuery, request.Volume, request.UserId, currencyId); err != nil {
		return err
	}

	return nil
}

func (p *PostgresProvider) getCurrencyNameByProtoCurrency(currency proto.Currency) (string, error) {
	switch currency {
	case proto.Currency_CURRENCY_EUR:
		return eurCurrencyName, nil
	case proto.Currency_CURRENCY_UAH:
		return uahCurrencyName, nil
	case proto.Currency_CURRENCY_USD:
		return usdCurrencyName, nil
	default:
		return "", errors.New(notImplementedCurrencyErrorMsg)
	}
}

// --------- Unlock Balance ---------

func (p *PostgresProvider) UnlockBalance(request *proto.UnlockBalanceRequest) error {
	currencyId, err := p.getCurrencyId(request.Currency)
	if err != nil {
		return err
	}

	conn := p.GetNewConnection()
	defer conn.Close(context.Background())

	var balance, lockedBalance float64
	if err = conn.QueryRow(context.Background(), getBalancesQuery, request.UserId, currencyId).Scan(&balance, &lockedBalance); err != nil {
		return fmt.Errorf(userNotFoundErrMsg, request.UserId)
	}

	if lockedBalance-request.Volume < 0 {
		return errors.New(notEnoughBalanceToUnLockErrorMsg)
	}

	if _, err = conn.Exec(context.Background(), unlockBalanceQuery, request.Volume, request.UserId, currencyId); err != nil {
		return err
	}

	return nil
}
