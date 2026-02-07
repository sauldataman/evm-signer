package base

import (
	"evm-signer/pkg/logging"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type SignerConfig struct {
	Config *viper.Viper
	Rule   []byte
}

var (
	scfg        *SignerConfig
	logger      *logging.SugaredLogger
	ServiceName = "signer"
)

func init() {
	scfg = new(SignerConfig)
	var err error
	scfg.Config, err = initConfigFromLocalFile("config", "yaml")
	if err != nil {
		fmt.Println("missing config.yaml file")
		os.Exit(1)
	}

	logger = GetLogger("parse").Sugar()
}

func initConfigFromLocalFile(fileName, fileType string) (*viper.Viper, error) {
	vr := viper.New()
	vr.SetConfigName(fileName)
	vr.SetConfigType(fileType)
	vr.AddConfigPath("./conf")
	vr.AddConfigPath("../conf")
	vr.AddConfigPath("../../conf")
	err := vr.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return vr, err
}

func GetSignerConfig(ruleName string) *SignerConfig {
	logger.Infof("ruleName: %s", ruleName)

	ruleFileArr := strings.Split(ruleName, ".")
	if len(ruleFileArr) < 2 {
		fmt.Println("invalidate file name, eg. rule.json")
		os.Exit(1)
	}

	if ruleFileArr[1] != "json" {
		logger.Fatalf(".%s unsupported file type, only JSON rules are supported", ruleFileArr[1])
		return nil
	}

	// Try to load rule file from conf directories
	confPaths := []string{"./conf", "../conf", "../../conf"}
	var rule []byte
	var err error
	var loadedPath string

	for _, confPath := range confPaths {
		rulePath := confPath + "/" + ruleName
		rule, err = os.ReadFile(rulePath)
		if err == nil {
			loadedPath = rulePath
			break
		}
	}

	if err != nil {
		logger.Fatalf("failed to load rule file %s from conf directories: %s", ruleName, err.Error())
	}

	logger.Infof("loaded rule file from: %s", loadedPath)
	scfg.Rule = rule
	return scfg
}

func GetConfig() *SignerConfig {
	return scfg
}

func GetLogger(module string) *logging.Logger {
	logConfig := logging.GetLogConfig(GetConfig().Config)
	return logging.GetLogger(ServiceName, module, logConfig)
}
