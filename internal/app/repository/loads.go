package repository

import (
	"fmt"
	"web/internal/app/ds"
)

func (r *Repository) GetAllLoads() ([]ds.Loads, error) {
	var loads []ds.Loads

	err := r.db.Find(&loads).Error
	if err != nil {
		return nil, err
	}

	if len(loads) == 0 {
		return nil, fmt.Errorf("loads not found")
	}
	return loads, nil
}

func (r *Repository) SearchLoadsByName(title string) ([]ds.Loads, error) {
	var loads []ds.Loads
	err := r.db.Where("load_title ILIKE ?", "%"+title+"%").Find(&loads).Error // добавили условие
	if err != nil {
		return nil, err
	}
	return loads, nil
}

func (r *Repository) GetLoadByID(id int) (*ds.Loads, error) {
	var load ds.Loads
	err := r.db.First(&load, id).Error
	if err != nil {
		return nil, err
	}
	return &load, nil
}
