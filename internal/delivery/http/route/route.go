package route

import (
	"iot-server/internal/delivery/http"

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
	c.App.POST("/api/users", c.UserController.Register)
	c.App.POST("/api/users/login", c.UserController.Login)
}

func (c *RouteConfig) SetupAuthRoute() {
	v1 := c.App.Group("/api/v1", c.AuthMiddleware)

	sensor := v1.Group("/sensor")
	sensor.POST("/create", c.SensorController.CreateSensor)

	sensor.GET("/search/by-id", c.SensorController.SearchByCombinedId)
	sensor.GET("/search/by-time-range", c.SensorController.SearchByTimeRange)
	sensor.GET("/search/by-id-time-range", c.SensorController.SearchByIdAndTimeRange)

	sensor.DELETE("/delete/by-id", c.SensorController.DeleteByCombinedId)
	sensor.DELETE("/delete/by-time-range", c.SensorController.DeleteByTimeRange)
	sensor.DELETE("/delete/by-id-time-range", c.SensorController.DeleteByIdAndTimeRange)

	sensor.PATCH("/update/by-id", c.SensorController.UpdateByCombinedId)
	sensor.PATCH("/update/by-time-range", c.SensorController.UpdateByTimeRange)
	sensor.PATCH("/update/by-id-time-range", c.SensorController.UpdateByIdAndTimeRange)

	// user
	user := c.App.Group("/api/users", c.AuthMiddleware)
	user.POST("/logout", c.UserController.Logout)
}
