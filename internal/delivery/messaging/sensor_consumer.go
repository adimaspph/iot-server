package messaging

import (
	"context"
	"encoding/json"
	"iot-server/internal/model"
	"iot-server/internal/usecase"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

type SensorConsumer struct {
	UseCase *usecase.SensorUsecase
	Log     *logrus.Logger
}

func NewSensorConsumer(useCase *usecase.SensorUsecase, logger *logrus.Logger) *SensorConsumer {
	return &SensorConsumer{
		UseCase: useCase,
		Log:     logger,
	}
}

func (c *SensorConsumer) SensorMQTTHandler(_ mqtt.Client, msg mqtt.Message) {

	var req model.CreateSensorRequest
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		c.Log.WithFields(logrus.Fields{
			"topic":   msg.Topic(),
			"payload": string(msg.Payload()),
		}).WithError(err).Warn("MQTT: invalid JSON")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.UseCase.Create(ctx, &req)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"topic":        msg.Topic(),
			"id1":          req.ID1,
			"id2":          req.ID2,
			"sensor_type":  req.SensorType,
			"sensor_value": req.SensorValue,
			"timestamp":    req.Timestamp,
		}).WithError(err).Error("MQTT: create failed")
		return
	}

	c.Log.WithFields(logrus.Fields{
		"topic":        msg.Topic(),
		"id1":          resp.ID1,
		"id2":          resp.ID2,
		"sensor_type":  resp.SensorType,
		"sensor_value": resp.SensorsRecords[0].SensorValue,
		"timestamp":    resp.SensorsRecords[0].Timestamp,
	}).Info("MQTT: sensor data created")
}
