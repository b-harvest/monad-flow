package util

import "sort"

type OrderedSet []uint16

func NewOrderedSet() OrderedSet {
	return nil
}

func (s OrderedSet) find(val uint16) (index int, found bool) {
	index = sort.Search(len(s), func(i int) bool {
		return s[i] >= val
	})
	if index < len(s) && s[index] == val {
		return index, true
	}
	return index, false
}

func (s *OrderedSet) Append(val uint16) {
	*s = append(*s, val)
}

func (s *OrderedSet) Insert(val uint16) bool {
	index, found := s.find(val)
	if found {
		return false
	}

	*s = append(*s, 0)
	copy((*s)[index+1:], (*s)[index:])
	(*s)[index] = val
	return true
}

func (s *OrderedSet) Remove(val uint16) bool {
	index, found := s.find(val)
	if !found {
		return false
	}
	*s = append((*s)[:index], (*s)[index+1:]...)
	return true
}

func (s *OrderedSet) InsertOrRemove(val uint16) {
	index, found := s.find(val)
	if found {
		*s = append((*s)[:index], (*s)[index+1:]...)
	} else {
		*s = append(*s, 0)
		copy((*s)[index+1:], (*s)[index:])
		(*s)[index] = val
	}
}

func (s OrderedSet) First() (uint16, bool) {
	if len(s) == 0 {
		return 0, false
	}
	return s[0], true
}

func (s OrderedSet) Values() []uint16 {
	return s
}
