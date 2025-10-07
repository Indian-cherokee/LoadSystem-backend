package ds

type LoadToCalculation struct {
	ID            uint `gorm:"primaryKey;column:id"`
	LoadSessionID uint `gorm:"column:load_session_id;not null"`
	LoadID        uint `gorm:"column:load_id;not null"`
	Area          *int `gorm:"column:area"`

	Load Loads `gorm:"foreignKey:LoadID;references:ID"`
}

func (LoadToCalculation) TableName() string { return "load_to_calculations" }
