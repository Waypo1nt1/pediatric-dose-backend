package ds

type DrugLike struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;uniqueIndex:idx_drug_likes_user_drug"`
	DrugID uint `gorm:"not null;uniqueIndex:idx_drug_likes_user_drug"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	Drug Drug `gorm:"foreignKey:DrugID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}
