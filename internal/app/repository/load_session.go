package repository

import (
	"errors"
	"web/internal/app/ds"
)

func (r *Repository) GetDraftLoadSession(userID uint) (*ds.LoadSession, error) {
	var load_session ds.LoadSession

	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&load_session).Error
	if err != nil {
		return nil, err
	}
	return &load_session, nil
}

func (r *Repository) CreateLoadSession(load_session *ds.LoadSession) error {
	return r.db.Create(load_session).Error
}

func (r *Repository) AddLoadToLoadSession(load_sessionID, loadID uint) error {
	var count int64

	r.db.Model(&ds.LoadToCalculation{}).Where("load_session_id = ? AND load_id = ?", load_sessionID, loadID).Count(&count)
	if count > 0 {
		return errors.New("load already in calculation session")
	}

	link := ds.LoadToCalculation{
		LoadSessionID: load_sessionID,
		LoadID:        loadID,
		Area:          nil,
	}
	return r.db.Create(&link).Error
}

func (r *Repository) GetLoadSessionWithLoads(load_sessionID uint) (*ds.LoadSession, error) {
	var load_session ds.LoadSession

	err := r.db.Preload("LoadsLink").Preload("LoadsLink.Load").First(&load_session, load_sessionID).Error
	if err != nil {
		return nil, err
	}

	if load_session.Status == ds.StatusDeleted {
		return nil, errors.New("load session page not found or has been deleted")
	}

	return &load_session, nil
}

func (r *Repository) LogicallyDeleteLoadSession(load_sessionID uint) error {
	result := r.db.Exec("UPDATE load_sessions SET status = ? WHERE id = ?", ds.StatusDeleted, load_sessionID)
	return result.Error
}
