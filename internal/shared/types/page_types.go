package types

import "github.com/lokeam/bravo-kilo/internal/shared/core"

type PageQueryParams struct {
	UserID int              `json:"userID" validate:"required"`
	Domain core.DomainType  `json:"domain" validate:"required,oneof=books games movies"`
}

type LibraryResponse struct {
	RequestID   string           `json:"requestId"`
	Data        *LibraryPageData `json:"data"`
	Source      string           `json:"source"`
}

type HomeResponse struct {
	RequestID    string        `json:"requestId"`
	Data         *HomePageData  `json:"data"`
	Source       string        `json:"source"`
}

