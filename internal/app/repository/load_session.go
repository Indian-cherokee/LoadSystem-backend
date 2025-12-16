package repository

import (
	"errors"
	"time"
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

func (r *Repository) GetLoadSessionsFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.LoadSession, error) {
	var sessions []ds.LoadSession
	// Исключаем только удаленные по умолчанию
	query := r.db.Where("status != ?", ds.StatusDeleted)

	// Черновики показываем только если они были явно сохранены (room_type IS NOT NULL)
	// Или если запрашиваются явно через фильтр status=draft
	if status == "draft" {
		// При явном запросе черновиков показываем все, включая несохраненные
	} else {
		// Для всех остальных случаев (включая "все статусы" когда status == "")
		// Исключаем черновики, которые еще не были сохранены (room_type IS NULL)
		// Это означает, что пользователь еще не нажал "Сохранить заявку"
		// При "все статусы" будут показаны только сохраненные черновики (room_type IS NOT NULL)
		query = query.Where("NOT (status = ? AND (room_type IS NULL OR room_type = ''))", ds.StatusDraft)
	}

	if !isModerator {
		query = query.Where("creator_id = ?", userID)
	}

	if status != "" {
		var statusInt int
		switch status {
		case "draft":
			statusInt = ds.StatusDraft
		case "formed":
			statusInt = ds.StatusFormed
		case "completed":
			statusInt = ds.StatusCompleted
		case "rejected":
			statusInt = ds.StatusRejected
		default:
			return sessions, nil
		}
		query = query.Where("status = ?", statusInt)
	}
	if from != "" {
		if fromTime, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("creation_date >= ?", fromTime)
		}
	}
	if to != "" {
		if toTime, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("creation_date <= ?", toTime)
		}
	}

	err := query.Preload("Creator").Preload("Moderator").Find(&sessions).Error
	return sessions, err
}

func (r *Repository) UpdateLoadSessionUserFields(id uint, req ds.LoadSessionUpdateRequest) error {
	updates := make(map[string]interface{})
	if req.RoomType != nil {
		updates["room_type"] = *req.RoomType
	}

	return r.db.Model(&ds.LoadSession{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) FormLoadSession(id, userID uint) error {
	// Проверяем, что сессия принадлежит пользователю и в статусе черновика
	var session ds.LoadSession
	if err := r.db.Where("id = ? AND creator_id = ? AND status = ?", id, userID, ds.StatusDraft).First(&session).Error; err != nil {
		return errors.New("сессия не найдена или не может быть сформирована")
	}

	// Проверяем, что в сессии есть нагрузки
	var count int64
	r.db.Model(&ds.LoadToCalculation{}).Where("load_session_id = ?", id).Count(&count)
	if count == 0 {
		return errors.New("нельзя сформировать пустую сессию")
	}

	// Обновляем статус и дату формирования
	now := time.Now()
	return r.db.Model(&session).Updates(map[string]interface{}{
		"status":       ds.StatusFormed,
		"forming_date": &now,
	}).Error
}

func (r *Repository) ResolveLoadSession(id, moderatorID uint, action string) error {
	var session ds.LoadSession
	if err := r.db.Where("id = ? AND status = ?", id, ds.StatusFormed).First(&session).Error; err != nil {
		return errors.New("сессия не найдена или не может быть обработана")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"moderator_id":    moderatorID,
		"completion_date": &now,
	}

	switch action {
	case "complete":
		updates["status"] = ds.StatusCompleted
		// total_load будет рассчитан асинхронным сервисом и обновлен позже

	case "reject":
		updates["status"] = ds.StatusRejected
	default:
		return errors.New("неверное действие")
	}

	return r.db.Model(&session).Updates(updates).Error
}

func (r *Repository) RemoveLoadFromSession(sessionID, loadID uint) error {
	return r.db.Where("load_session_id = ? AND load_id = ?", sessionID, loadID).Delete(&ds.LoadToCalculation{}).Error
}

func (r *Repository) UpdateLoadToCalculation(sessionID, loadID uint, updateData ds.LoadToCalculation) error {
	return r.db.Model(&ds.LoadToCalculation{}).Where("load_session_id = ? AND load_id = ?", sessionID, loadID).Updates(updateData).Error
}

// UpdateLoadSessionTotalLoad обновляет поле total_load для сессии
func (r *Repository) UpdateLoadSessionTotalLoad(sessionID uint, totalLoad float64) error {
	return r.db.Model(&ds.LoadSession{}).Where("id = ?", sessionID).Update("total_load", totalLoad).Error
}
