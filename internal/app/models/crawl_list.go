package models

import "time"

type CrawlListResponse struct {
	ID              uint      `json:"id"`
	URL             string    `json:"url"`
	PageTitle       string    `json:"page_title"`
	CreatedAt       time.Time `json:"created_at"`
	HTMLVersion     string    `json:"html_version"`
	Headings        HeadingCounts `json:"headings"`
	InternalLinks   int       `json:"internal_links"`
	ExternalLinks   int       `json:"external_links"`
	InaccessibleLinks StringSlice `json:"inaccessible_links"`
	HasLoginForm    bool      `json:"has_login_form"`
}
