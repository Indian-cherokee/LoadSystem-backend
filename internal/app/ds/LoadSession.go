package ds

import "time"

type LoadSession struct {
	ID             uint       `gorm:"primaryKey;column:id"`
	Status         int        `gorm:"column:status;not null"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	CreatorID      uint       `gorm:"column:creator_id;not null"`
	ModeratorID    *uint      `gorm:"column:moderator_id"`
	RoomType       *string    `gorm:"column:room_type"`
	FormingDate    *time.Time `gorm:"column:forming_date"`
	CompletionDate *time.Time `gorm:"column:completion_date"`

	Creator   Users               `gorm:"foreignKey:CreatorID"`
	Moderator *Users              `gorm:"foreignKey:ModeratorID"`
	LoadsLink []LoadToCalculation `gorm:"foreignKey:LoadSessionID"`
}

func (LoadSession) TableName() string { return "load_sessions" }
