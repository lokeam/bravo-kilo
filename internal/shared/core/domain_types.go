package core

import "context"

type DomainType string

const (
    BookDomainType  DomainType = "books"
    GameDomainType  DomainType = "games"
    MovieDomainType DomainType = "movies"
)

type DomainHandler interface {
    GetType() DomainType
    GetLibraryItems(ctx context.Context, userID int) ([]LibraryItem, error)
    GetMetadata() (DomainMetadata, error)
}

type PageType string

const (
    LibraryPage PageType = "library"
    HomePage    PageType = "home"
)

type LibraryItem struct {
    ID          int         `json:"id"`
    Title       string      `json:"title"`
    Type        DomainType  `json:"type"`
    DateAdded   string      `json:"dateAdded"`
    LastUpdated string      `json:"lastUpdated"`
}

type DomainMetadata struct {
    DomainType DomainType
    Label      string
}

// Helper fns for type checking
func (d DomainType) IsValid() bool {
    switch d {
        case BookDomainType, GameDomainType, MovieDomainType:
            return true
        default:
            return false
    }
}

func (p PageType) IsValid() bool {
    switch p {
    case LibraryPage, HomePage:
        return true
    default:
        return false
    }
}