package hin

import (
	"bytes"
	"fmt"
	"github.com/rs/xid"
	"log"
)

type HID struct {
	id xid.ID
}

func (h HID) String() string {
	return h.id.String()
}

func (h HID) IsNil() bool {
	return h.id.IsNil() || h.id.IsZero()
}

// MarshalJSON implements json.Marshal.
func (h *HID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", h.String())), nil
}

func (h *HID) UnmarshalJSON(b []byte) error {
	log.Printf(string(b))
	b = bytes.Trim(b, "\"")
	id, _ := xid.FromString(string(b))
	*h = HID{id: id}
	return nil
}

func NewID() HID {
	return HID{
		id: xid.New(),
	}
}

func ParseID(str string) HID {
	id, err := xid.FromString(str)
	if err != nil {
		log.Fatalf("parse %s to xid error, %v", str, err)
	}
	return HID{
		id: id,
	}
}
