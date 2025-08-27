package route

import (
	"iot-server/internal/delivery/http"
	"iot-server/internal/delivery/http/middleware"
	"iot-server/internal/entity"

	"github.com/labstack/echo/v4"
)

type RouteConfig struct {
	App              *echo.Echo
	SensorController *http.SensorController
	UserController   *http.UserController
	AuthMiddleware   echo.MiddlewareFunc
}

func (c *RouteConfig) Setup() {
	c.SetupGuestRoute()
	c.SetupAuthRoute()
}

func (c *RouteConfig) SetupGuestRoute() {
	c.App.POST("/api/users/login", c.UserController.Login)
}

func (c *RouteConfig) SetupAuthRoute() {
	v1 := c.App.Group("/api/v1", c.AuthMiddleware)

	sensor := v1.Group("/sensor")
	// Authenticated
	sensor.GET("/search/by-id", c.SensorController.SearchByCombinedId)
	sensor.GET("/search/by-time-range", c.SensorController.SearchByTimeRange)
	sensor.GET("/search/by-id-time-range", c.SensorController.SearchByIdAndTimeRange)

	// Admin-only (mutations)
	admin := sensor.Group("", middleware.RequireRoles(entity.RoleAdmin))
	admin.POST("/create", c.SensorController.CreateSensor)
	admin.DELETE("/delete/by-id", c.SensorController.DeleteByCombinedId)
	admin.DELETE("/delete/by-time-range", c.SensorController.DeleteByTimeRange)
	admin.DELETE("/delete/by-id-time-range", c.SensorController.DeleteByIdAndTimeRange)
	admin.PATCH("/update/by-id", c.SensorController.UpdateByCombinedId)
	admin.PATCH("/update/by-time-range", c.SensorController.UpdateByTimeRange)
	admin.PATCH("/update/by-id-time-range", c.SensorController.UpdateByIdAndTimeRange)

	// Authenticated
	user := c.App.Group("/api/users", c.AuthMiddleware)
	user.POST("", c.UserController.Register, middleware.RequireRoles(entity.RoleAdmin))
	user.GET("/logout", c.UserController.Logout)
}
