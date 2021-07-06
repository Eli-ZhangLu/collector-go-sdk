package common

// Set is a string set
type Set struct {
	data map[string]bool
}

func NewSet() Set {
	return Set{
		data: make(map[string]bool),
	}
}

func (s *Set) Size() int {
	return len(s.data)
}

func (s *Set) Copy() *Set {
	dst := &Set{
		data: make(map[string]bool),
	}
	for key, value := range s.data {
		dst.data[key] = value
	}
	return dst
}

func (s *Set) Insert(key string) {
	s.data[key] = true
}

func (s *Set) Delete(key string) {
	delete(s.data, key)
}

func (s *Set) Exist(key string) bool {
	_, exist := s.data[key]
	if exist {
		return true
	}
	return false
}

func (s *Set) Keys() map[string]bool {
	return s.data
}

// InterfaceSet
type InterfaceSet struct {
	data map[interface{}]bool
}

func NewInterfaceSet() InterfaceSet {
	return InterfaceSet{
		data: make(map[interface{}]bool),
	}
}

func (s *InterfaceSet) Size() int {
	return len(s.data)
}

func (s *InterfaceSet) Copy() *InterfaceSet {
	dst := &InterfaceSet{
		data: make(map[interface{}]bool),
	}
	for key, value := range s.data {
		dst.data[key] = value
	}
	return dst
}

func (s *InterfaceSet) Insert(key interface{}) {
	s.data[key] = true
}

func (s *InterfaceSet) Delete(key interface{}) {
	delete(s.data, key)
}

func (s *InterfaceSet) Exist(key interface{}) bool {
	_, exist := s.data[key]
	if exist {
		return true
	}
	return false
}

func (s *InterfaceSet) Keys() map[interface{}]bool {
	return s.data
}
