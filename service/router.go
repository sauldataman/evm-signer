package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) GetRouter() http.Handler {
	router := gin.Default()
	router.GET("/ping", s.Pong)
	router.POST("/v1/sign/transaction", s.GetSign)
	router.POST("/v1/sign/eip712", s.GetSign712)
	router.POST("/v1/sign/message", s.GetSignMessage)
	router.POST("/v1/address", s.GetAddress)
	return router
}
