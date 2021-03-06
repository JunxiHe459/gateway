package router

import (
	"github.com/JunxiHe459/gateway/controller"
	"github.com/JunxiHe459/gateway/docs"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"log"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationurl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationurl https://example.com/oauth/authorize
// @scope.admin Grants read and write access to administrative information

// @x-extension-openapi {"example": "value on a json format"}

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	router.Use(middlewares...)

	// Set up swagger
	docs.SwaggerInfo.Title = lib.GetStringConf("base.swagger.title")
	docs.SwaggerInfo.Description = lib.GetStringConf("base.swagger.desc")
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = lib.GetStringConf("base.swagger.host")
	docs.SwaggerInfo.BasePath = lib.GetStringConf("base.swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	ginSwagger.WrapHandler(swaggerFiles.Handler)

	// ping
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// AdminGroup ????????????
	adminGroup := router.Group("/admin")
	redis, err := sessions.NewRedisStore(10, "tcp", "localhost:6379", "Karlhe459!", []byte("secret"))
	if err != nil {
		print("NewRedisStore Error: ", err.Error())
		log.Fatalf("NewRedisStore Error: %v", err.Error())
	}
	// ??? adminGroup ???????????????
	adminGroup.Use(
		sessions.Sessions("AdminSession", redis),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ParamValidationMiddleware(),
	)
	// ??? adminGroup ?????? Admin Log in ??? /admin/login ????????????
	controller.RegiterAdmin(adminGroup)
	// ???????????????
	//adminLogin := controller.AdminLoginController{}
	//adminGroup.POST("/login", adminLogin.AdminLogin)

	// ??? adminInfoGroup ?????? Admin Info	 ??? /admin/info ????????????
	adminInfoGroup := router.Group("/admin/info")
	if err != nil {
		print("NewRedisStore Error: ", err.Error())
		log.Fatalf("NewRedisStore Error: %v", err.Error())
	}
	// ??? adminGroup ???????????????
	adminInfoGroup.Use(
		sessions.Sessions("AdminSession", redis),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		// ???????????? session ??????????????????
		middleware.SessionAuthMiddleware(),
		middleware.ParamValidationMiddleware(),
	)
	controller.RegiterAdminInfo(adminInfoGroup)

	// Service
	serviceGroup := router.Group("/service")
	if err != nil {
		print("NewRedisStore Error: ", err.Error())
		log.Fatalf("NewRedisStore Error: %v", err.Error())
	}
	serviceGroup.Use(
		sessions.Sessions("AdminSession", redis),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.SessionAuthMiddleware(),
		middleware.ParamValidationMiddleware(),
	)
	controller.RegisterService(serviceGroup)

	renterGroup := router.Group("/renter")
	renterGroup.Use(
		sessions.Sessions("AdminSession", redis),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.SessionAuthMiddleware(),
		middleware.ParamValidationMiddleware(),
	)
	controller.RenterRegister(renterGroup)

	dashboardGroup := router.Group("/dashboard")
	dashboardGroup.Use(
		sessions.Sessions("AdminSession", redis),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.SessionAuthMiddleware(),
		middleware.ParamValidationMiddleware(),
	)
	controller.DashboardRegister(dashboardGroup)

	return router

}
