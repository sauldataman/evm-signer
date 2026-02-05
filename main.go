package main

import (
	"context"
	"cs-evm-signer/base"
	"cs-evm-signer/service"
	"github.com/CoinSummer/go-base/logging"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	port     int
	ruleFile string
	logger   *logging.SugaredLogger
)

func init() {
	logger = base.GetLogger("signer").Sugar()
}

var rootCmd = &cobra.Command{
	Use: "signer",
	Long: `
Sodium platform transaction signature machine. 
You can decide which transactions are allowed to be signed through the rule.json file`,
}

func main() {
	startCmd.PersistentFlags().IntVarP(&port, "port", "p", 80, "specify the port on which the signer run")
	startCmd.PersistentFlags().StringVarP(&ruleFile, "rule", "r", "rule.json", "rule file name, eg. rule.json")
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(startCmd)
	_ = rootCmd.Execute()
}

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "signer start",
	Example: "./signer start --port 8080",
	Run: func(cmd *cobra.Command, args []string) {
		service.SetLogger(base.GetLogger("signer").Sugar())
		signerConfig := base.GetSignerConfig(ruleFile)
		accountForAddr, accountForIndex, iAccount := service.GetAccount(signerConfig)
		chains, err := service.GetChain(signerConfig)
		if err != nil {
			logger.Errorf("get chain fail: %s", err.Error())
			return
		}

		// init rule
		rules, err := service.GetRuleConfig(signerConfig)
		if err != nil {
			logger.Errorf("get rule fail: %s", err.Error())
			return
		}

		authConfig := service.GetAuthConfig(signerConfig)
		ipList := strings.Split(authConfig.IP, ",")
		whitelist := service.GetIpWhiteList(ipList)
		svc, err := service.New(iAccount, whitelist)
		if err != nil {
			logger.Errorf("service initialization fail: %s", err.Error())
			return
		}

		svc.SetAccountMap(accountForAddr)
		svc.SetAccountListMap(accountForIndex)
		svc.SetChainMap(chains)
		svc.SetRules(rules)

		httpConfig := service.GetHttpConfig(signerConfig)
		_port := 0
		if port != 80 {
			_port = port
		} else {
			_port = httpConfig.Port
		}
		router := svc.GetRouter()
		s := &http.Server{
			Addr:           ":" + strconv.Itoa(_port),
			Handler:        router,
			MaxHeaderBytes: 1 << 20,
		}

		go func() {
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("s.ListenAndServe err: %v", err)
			}
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shuting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}

		log.Println("Server exiting")
	},
}
