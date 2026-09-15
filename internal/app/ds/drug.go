package ds

import (
	"database/sql"
	"time"
)

const (
	DrugStatusDraft     = "draft"
	DrugStatusPublished = "published"
	DrugStatusDeleted   = "deleted"
)

type Drug struct {
	ID                     uint            `gorm:"primaryKey"`
	DrugName               string          `gorm:"type:varchar(100);not null"`
	ShortInfo              sql.NullString  `gorm:"type:varchar(500)"`
	DrugStatus             string          `gorm:"type:varchar(15);not null"`
	ImageURL               sql.NullString  `gorm:"type:varchar(255)"`
	VideoURL               sql.NullString  `gorm:"type:varchar(255)"`
	RecommendedAdultDoseMg sql.NullFloat64 `gorm:"type:numeric(8,2)"`
	MaxDailyDoseMg         sql.NullFloat64 `gorm:"type:numeric(8,2)"`
	CreatedAt              time.Time       `gorm:"not null"`
	CreatorID              uint            `gorm:"not null;uniqueIndex:idx_drugs_single_draft,where:drug_status = 'draft'"`
	PublishedAt            sql.NullTime

	Creator User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}
