package main

import (
	"fmt"
	"strings"

	appd "github.com/stone-payments/go-appdynamics"

	"github.com/gin-gonic/gin"
)

func AppDynamicsMiddleware() func(*gin.Context) {

	cfg := appd.Config{}

	cfg.AppName = "TestAppD"
	// cfg.TierName = conf.Name
	// cfg.NodeName = conf.Hostname
	// cfg.Controller.Host = conf.AppDynamics.Host
	// cfg.Controller.Port = uint16(conf.AppDynamics.Port)
	// cfg.Controller.UseSSL = conf.AppDynamics.UseSSL
	// cfg.Controller.Account = conf.AppDynamics.Account
	// cfg.Controller.AccessKey = conf.AppDynamics.AccessKey
	cfg.InitTimeoutMs = 0

	if err := appd.InitSDK(&cfg); err != nil {
		fmt.Printf("Error initializing the AppDynamics SDK - %#v\n", err.Error())
	} else {
		fmt.Printf("Initialized AppDynamics SDK successfully\n")
	}

	return func(c *gin.Context) {
		route := c.Request.Method + " " + c.Request.URL.Path
		for k := range c.Params {
			p := c.Params[k].Value
			route = strings.Replace(route, p, "*", -1)
		}

		// Start a AppDynamics business transaction
		btHandle := appd.StartBT(route, "")
		for k := range c.Params {
			p := c.Params[k]
			appd.AddUserDataToBT(btHandle, p.Key, p.Value)
		}

		// Call pending handlers
		c.Next()

		status := c.Writer.Status()
		if status > 200 {
			if status > 500 {
				appd.AddBTError(btHandle, appd.APPD_LEVEL_ERROR, "500", true)
			} else {
				appd.AddBTError(btHandle, appd.APPD_LEVEL_WARNING, "500", false)
			}
		}

		// End and save the business transaction
		appd.EndBT(btHandle)
	}

}

func main() {
	r := gin.Default()
	r.Use(AppDynamicsMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/ping/:me", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run()
}
