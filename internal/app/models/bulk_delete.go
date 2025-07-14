package models

type BulkDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}
