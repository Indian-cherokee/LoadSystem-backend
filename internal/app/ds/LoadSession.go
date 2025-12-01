package ds

import "time"

type LoadSession struct {
	ID             uint       `gorm:"primaryKey;column:id"`
	Status         int        `gorm:"column:status;not null"`
	CreationDate   time.Time  `gorm:"column:creation_date;not null"`
	CreatorID      uint       `gorm:"column:creator_id;not null"`
	RoomType       *string    `gorm:"column:room_type"`
	ModeratorID    *uint      `gorm:"column:moderator_id"`
	FormingDate    *time.Time `gorm:"column:forming_date"`
	CompletionDate *time.Time `gorm:"column:completion_date"`

	Creator   Users               `gorm:"foreignKey:CreatorID"`
	Moderator *Users              `gorm:"foreignKey:ModeratorID"`
	LoadsLink []LoadToCalculation `gorm:"foreignKey:LoadSessionID"`
}

func (LoadSession) TableName() string { return "load_sessions" }
