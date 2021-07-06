package storage

import (
	"testing"
)

func Test_Local_Normal(t *testing.T) {
	var c Storage
	var err error

	c, err = NewLocalStorage("abc.db")
	if err != nil {
		t.Fatal(err)
	}

	key := "key"
	val := "value"
	if err = c.Set(key, val, 0); err != nil {
		t.Fatal(err)
	}

	v, err := c.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if v != val {
		t.Fatal("value is not correct")
	}

	err = c.Del(key)
	if err != nil {
		t.Fatal(err)
	}

	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}

	// clear db
	err = c.Destory()
	if err != nil {
		t.Fatal(err)
	}

}

func Test_Local_DoubleClose(t *testing.T) {
	var c Storage
	var err error

	c, err = NewLocalStorage("abc.db")
	if err != nil {
		t.Fatal(err)
	}

	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}

	// clear db
	err = c.Destory()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Local_NonExistValue(t *testing.T) {
	var c Storage
	var err error

	c, err = NewLocalStorage("abc.db")
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Get("not_existed")
	if err == nil {
		t.Fatal(err)
	}

	err = c.Del("not_existed")
	if err != nil {
		t.Fatal(err)
	}

	// clear db
	err = c.Destory()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Local_CoverSet(t *testing.T) {
	var c Storage
	var err error
	var v string

	c, err = NewLocalStorage("abc.db")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	key := "same"
	if err = c.Set(key, "v1", 0); err != nil {
		t.Fatal(err)
	}

	if err = c.Set(key, "v2", 0); err != nil {
		t.Fatal(err)
	}

	v, err = c.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if v != "v2" {
		t.Fatal(err)
	}

	// clear db
	err = c.Destory()
	if err != nil {
		t.Fatal(err)
	}
}
