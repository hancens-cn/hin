package hin

import "testing"

func TestCriteria(t *testing.T) {
	builder := Criteria("username = ? AND password = ? AND id != 1 OR id = m OR item != ? AND tt = 1", "hancens", "123456", "asb")
	t.Log(builder)

	bs := builder.Mgo()
	t.Log(bs)
}
