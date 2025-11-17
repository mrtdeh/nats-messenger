package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Serve(h http.Handler) error {

	endPoint := "0.0.0.0:3000" // Debug Mode
	if !viper.GetBool(DEBUG) { // Release Mode
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
		endPoint = "localhost:3000"
	}

	server := &http.Server{
		Addr:    endPoint,
		Handler: h,
	}

	log.Printf("[info] start http server listening %s", endPoint)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
