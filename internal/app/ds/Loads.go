package ds

type Loads struct {
	ID                     uint    `gorm:"primaryKey;column:id"`
	LoadTitle              string  `gorm:"column:load_title;size:255;not null"`
	LoadDescription        string  `gorm:"column:load_description;type:text"`
	LoadImage              *string `gorm:"column:load_image;size:1024"`
	Normative              float64 `gorm:"column:normative"`
	LoadCategory           string  `gorm:"column:load_category;size:255"`
	ReliabilityCoefficient float64 `gorm:"column:reliability_coefficient"`
	Status                 *bool   `gorm:"column:status"`

	CalcLinks []LoadToCalculation `gorm:"foreignKey:LoadID;references:ID"`
}

func (Loads) TableName() string { return "loads" }
