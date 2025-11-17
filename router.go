package main

// func InitRouter(c *Client) http.Handler {

// 	if viper.GetBool("http_debug_mode") {
// 		gin.SetMode(gin.DebugMode)
// 	} else {
// 		gin.SetMode(gin.ReleaseMode)
// 		gin.DefaultWriter = io.Discard
// 	}

// 	r := gin.New()
// 	r.Use(gin.Logger())
// 	r.Use(gin.Recovery())

// 	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 	r.GET("/health/system", h.CheckSystemHealth)
// 	r.GET("/health/services", h.CheckServicesHealth)

// 	r.GET("/nodes", h.ListNodes)
// 	r.DELETE("/nodes/:id", h.DeleteNode)
// 	r.GET("/update/offlines", h.ListUpdates)
// 	r.POST("/update/offline", h.OfflineUpdate)

// 	r.POST("/service/action", h.ExecAction)

// 	return r
// }
