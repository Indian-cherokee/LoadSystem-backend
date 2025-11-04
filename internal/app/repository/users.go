package repository

import (
	"web/internal/app/ds"

	"golang.org/x/crypto/bcrypt"
)

func (r *Repository) CreateUser(user *ds.Users) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (*ds.Users, error) {
	var user ds.Users
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByUsername(username string) (*ds.Users, error) {
	var user ds.Users
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(id uint, req ds.UserUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password"] = string(hashedPassword)
	}

	return r.db.Model(&ds.Users{}).Where("id = ?", id).Updates(updates).Error
}
