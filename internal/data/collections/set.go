package collections

var exists = struct{}{}

// HashSet represents a collection of unique strings
type Set struct{
	m map[string]struct{}
}

// Creates and Returns a new Set
func NewSet() *Set {
	return &Set{m:make(map[string]struct{})}
}

// Inserts a new element into Set
func (s *Set) Add(value string) {
	s.m[value] = exists
}

// Deletes an element from Set
func (s *Set) Delete(value string) {
	delete(s.m, value)
}

// Checks if an element exists in Set
func (s *Set) Has(value string) bool {
	_, exists := s.m[value]
	return exists
}

// Returns the number of elements in Set
func (s *Set) Size() int {
	return len(s.m)
}

// Removes all elements from Set
func (s *Set) Clear() {
	s.m = make(map[string]struct{})
}

// Returns all elements in set as a slice of strings
func (s *Set) Elements() []string {
	elements := make([]string, 0, len(s.m))
	for key := range s.m {
		elements = append(elements, key)
	}

	return elements
}