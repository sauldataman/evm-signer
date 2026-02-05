package service

import (
	"cs-evm-signer/service/account"
	"cs-evm-signer/service/rules"
	"cs-evm-signer/types"
	"fmt"
	"github.com/CoinSummer/go-base/logging"
	"github.com/gin-gonic/gin"
	"strings"
	"sync"
)

var logger *logging.SugaredLogger

type Service struct {
	lock            sync.Mutex
	accountsForAddr map[string]*types.Account
	accountForIndex map[int64]*types.Account
	iAccount        account.IAccount
	chains          map[uint64]*ChainConfig
	whitelists      map[string]struct{}
	rules           rules.Rules
}

func SetLogger(_logger *logging.SugaredLogger) {
	logger = _logger
	rules.SetLogger(logger)
	account.SetLogger(logger)
}

func (s *Service) SetAccountMap(_accountMap map[string]*types.Account) {
	s.accountsForAddr = _accountMap
}

func (s *Service) SetAccountListMap(_accountListMap map[int64]*types.Account) {
	s.accountForIndex = _accountListMap
}

func (s *Service) GetAccount(accountAddr string) (account *types.Account, ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	account, ok = s.accountsForAddr[strings.ToLower(accountAddr)]
	if !ok {
		return nil, ok
	}
	if account.PriKey == nil && account.HsmClient == nil {
		return nil, false
	}
	return
}

func (s *Service) GetAccountList(index int64) (account *types.Account, ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.iAccount.Account().GetAccount().Index(s.accountForIndex, index)
}

func (s *Service) SetRules(rs rules.Rules) {
	s.rules = rs
	s.rules.Init()
}

// New create new monitor service
func New(iAccount account.IAccount, whitelists map[string]struct{}) (srv *Service, err error) {
	srv = &Service{
		iAccount:   iAccount,
		whitelists: whitelists,
	}
	return srv, nil
}

func (s *Service) ServiceName() string {
	return "signer"
}

func (s *Service) Pong(ctx *gin.Context) {
	if err := s.CheckIp(ctx); err != nil {
		_msg := fmt.Sprintf("ip: %s illegal", ctx.ClientIP())
		ReturnError(ctx, IllegalAccess, _msg)
		return
	}
	ReturnSuccess(ctx, "pong")
	return
}
