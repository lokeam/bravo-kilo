package domains

import "fmt"

type BookDomainError struct {
    Source string
    Err    error
}

func (e *BookDomainError) Error() string {
    return fmt.Sprintf("book domain error in %s: %v", e.Source, e.Err)
}