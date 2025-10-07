package ds

import "time"

type LoadSession struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Status    int       `gorm:"column:status;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	CreatorID uint      `gorm:"column:creator_id;not null"`

	// Связи
	Creator   Users               `gorm:"foreignKey:CreatorID"`
	LoadsLink []LoadToCalculation `gorm:"foreignKey:LoadSessionID"`
}

func (LoadSession) TableName() string { return "load_sessions" }
