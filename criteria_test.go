package hin

import "testing"

func TestCriteria(t *testing.T) {
	builder := Criteria("username = ? AND password = ? AND id != 1", "hancens", "123456")
	t.Log(builder)

	bs := builder.Mgo()
	t.Log(bs)
}
