package converter

import (
	"iot-subscriber/internal/entity"
	"iot-subscriber/internal/model"
)

func SensorToResponse(sensor *entity.Sensor) *model.SensorResponse {

	sensorRecords := make([]model.SensorRecord, 0)

	for _, record := range sensor.Records {
		sensorRecords = append(sensorRecords, model.SensorRecord{
			SensorValue: record.SensorValue,
			Timestamp:   record.Timestamp,
		})
	}

	return &model.SensorResponse{
		ID1:            sensor.ID1,
		ID2:            sensor.ID2,
		SensorType:     sensor.SensorType,
		SensorsRecords: sensorRecords,
	}
}
