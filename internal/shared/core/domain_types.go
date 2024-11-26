package core

type DomainType string

const (
    BookDomainType  DomainType = "books"
    GameDomainType  DomainType = "games"
    MovieDomainType DomainType = "movies"
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