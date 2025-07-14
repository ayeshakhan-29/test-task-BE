package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
)

type CrawlRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type HeadingCounts struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

// GORM Scan interface
func (h *HeadingCounts) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, h)
}

// GORM Value interface
func (h HeadingCounts) Value() (driver.Value, error) {
	return json.Marshal(h)
}

type CrawlResult struct {
	gorm.Model
	URL               string       `json:"url" gorm:"type:varchar(2000);not null"`
	HTMLVersion       string       `json:"html_version" gorm:"size:50"`
	PageTitle         string       `json:"page_title" gorm:"type:text"`
	Headings          HeadingCounts `json:"headings" gorm:"type:JSON"`
	InternalLinks     int          `json:"internal_links" gorm:"default:0"`
	ExternalLinks     int          `json:"external_links" gorm:"default:0"`
	InaccessibleLinks int          `json:"inaccessible_links" gorm:"default:0"`
	HasLoginForm      bool         `json:"has_login_form" gorm:"default:false"`
	UserID            uint64       `json:"user_id" gorm:"index;not null"`
	User              User         `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
