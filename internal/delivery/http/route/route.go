package route

import (
	"iot-subscriber/internal/delivery/http"

	"github.com/labstack/echo/v4"
)

type RouteConfig struct {
	App              *echo.Echo
	SensorController *http.SensorController
}

func (c *RouteConfig) Setup() {
	c.SetupAuthRoute()
}

func (c *RouteConfig) SetupAuthRoute() {

	v1 := c.App.Group("/api/v1")

	sensor := v1.Group("/sensor")
	sensor.POST("/create", c.SensorController.CreateSensor)
	sensor.GET("/search/by-id", c.SensorController.SearchByCombinedId)
	sensor.GET("/search/by-time-range", c.SensorController.SearchByTimeRange)
}
