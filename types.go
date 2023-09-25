package hin

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// H is a shortcut for map[string]any
type H map[string]any

type UnixTime time.Time

// MarshalJSON implements json.Marshal.
func (t *UnixTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", time.Time(*t).UnixMilli())
	return []byte(stamp), nil
}

func (t *UnixTime) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	if ms, err := strconv.ParseInt(string(b), 10, 64); err != nil {
		return err
	} else {
		*t = UnixTime(time.UnixMilli(ms))
	}
	return nil
}

func NewUnixTime(t time.Time) *UnixTime {
	tm := UnixTime(t)
	return &tm
}

type BaseModel struct {
	ID        string     `json:"id" bson:"_id,minsize"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" bson:"deleted_at,omitempty"`
}

type IdentityQuery struct {
	ID string `uri:"id" binding:"required" criteria:"eq,_id"`
}

type PagingQuery struct {
	Page  int64 `form:"page,default=0"`
	Count int64 `form:"count,default=20"`
}

type PagingDTO struct {
	Page  int64 `json:"page"`
	Count int64 `json:"count"`
	Total int64 `json:"total"`
	Items any   `json:"items"`
}
