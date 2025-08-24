package converter

import (
	"fmt"
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

func SensorRecordsToResponse(records []entity.SensorRecord) []model.SensorResponse {
	if len(records) == 0 {
		return []model.SensorResponse{}
	}

	// Map key: "id1|id2"
	grouped := make(map[string]*model.SensorResponse)

	for _, rec := range records {
		key := rec.Sensor.ID1 + "|" +
			fmt.Sprintf("%d", rec.Sensor.ID2)

		// Initialize if not exists
		if _, exists := grouped[key]; !exists {
			grouped[key] = &model.SensorResponse{
				ID1:        rec.Sensor.ID1,
				ID2:        rec.Sensor.ID2,
				SensorType: rec.Sensor.SensorType,
			}
		}

		// Append record to the sensor response
		grouped[key].SensorsRecords = append(grouped[key].SensorsRecords, model.SensorRecord{
			SensorValue: rec.SensorValue,
			Timestamp:   rec.Timestamp,
		})
	}

	// Convert map into slice
	responses := make([]model.SensorResponse, 0, len(grouped))
	for _, resp := range grouped {
		responses = append(responses, *resp)
	}

	return responses
}
