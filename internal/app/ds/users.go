package ds

type Users struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	Username  string `gorm:"unique;column:username;size:255;not null"`
	Password  string `gorm:"unique;column:password;size:255;not null"`
	Moderator bool   `gorm:"column:moderator;not null"`

	CalcSessions []LoadSession `gorm:"foreignKey:CreatorID"`
}

func (Users) TableName() string { return "users" }
