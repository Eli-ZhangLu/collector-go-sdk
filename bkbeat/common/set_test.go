package common

import (
	"testing"
)

func Test_Set(t *testing.T) {
	s := NewSet()

	// empty
	n := s.Size()
	if n != 0 {
		t.Error("set is not empty")
		t.Fail()
	}

	s.Insert("abc1")
	s.Insert("abc2")
	s.Insert("abc1")

	n = s.Size()
	if n != 2 {
		t.Error("set size is not 2")
		t.Fail()
	}

	s2 := s.Copy()
	if !s2.Exist("abc2") {
		t.Error("can not find abc2")
		t.Fail()
	}

	s2.Delete("abc2")

	if s2.Exist("abc2") {
		t.Error("should not exist abc2")
		t.Fail()
	}

	if s.Size() != 2 {
		t.Error("copy not work")
		t.Fail()
	}

	for key, _ := range s.Keys() {
		t.Log("key:", key)
	}
}

func Test_IntSet(t *testing.T) {
	s := NewInterfaceSet()

	// empty
	n := s.Size()
	if n != 0 {
		t.Error("set is not empty")
		t.Fail()
	}

	s.Insert(111)
	s.Insert(222)
	s.Insert(111)

	n = s.Size()
	if n != 2 {
		t.Error("set size is not 2")
		t.Fail()
	}

	s2 := s.Copy()
	if !s2.Exist(222) {
		t.Error("can not find 222")
		t.Fail()
	}

	s2.Delete(222)

	if s2.Exist(222) {
		t.Error("should not exist 222")
		t.Fail()
	}

	if s.Size() != 2 {
		t.Error("copy not work")
		t.Fail()
	}

	for key, _ := range s.Keys() {
		t.Log("key:", key)
	}
}
