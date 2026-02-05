package service

type ChainConfig struct {
	Name      string
	ChainType string `mapstructure:"chain_type"`
	ChainId   uint64 `mapstructure:"chain_id"`
}

func (s *Service) SetChainMap(am map[uint64]*ChainConfig) {
	s.chains = am
}

func (s *Service) GetChainConfig(chainID uint64) *ChainConfig {
	s.lock.Lock()
	defer s.lock.Unlock()

	chain, ok := s.chains[chainID]
	if !ok {
		return nil
	}
	return chain
}
