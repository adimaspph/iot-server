package converter

import (
	"iot-server/internal/entity"
	"iot-server/internal/model"
)

func UserToResponse(user *entity.User) *model.UserResponse {
	return &model.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		Role: user.Role,
	}
}

func UserToTokenResponse(user *entity.User, token string) *model.UserResponse {
	return &model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Token: token,
		Role:  user.Role,
	}
}
