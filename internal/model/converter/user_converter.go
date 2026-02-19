package converter

import (
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
)

func UserToResponse(user *entity.User, roles []string) *model.UserResponse {
	return &model.UserResponse{
		ID:            user.ID,
		FullName:      user.FullName,
		Email:         user.Email,
		Phone:         user.Phone,
		ProfileImage:  user.ProfileImage,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		Roles:         roles,
		CreatedAt:     user.CreatedAt.UnixMilli(),
	}
}
