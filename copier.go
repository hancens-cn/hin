package hin

import (
	"errors"
	"github.com/jinzhu/copier"
	"time"
)

var (
	_Int64 int64 = 0
)

func Copy(to any, from any) error {
	opts := copier.Option{
		Converters: []copier.TypeConverter{
			{
				SrcType: HID{},
				DstType: copier.String,
				Fn: func(from any) (any, error) {
					s, ok := from.(HID)
					if !ok {
						return nil, errors.New("src type not matching to hin.HID")
					}
					return s.String(), nil
				},
			},
			{
				SrcType: copier.String,
				DstType: HID{},
				Fn: func(from any) (any, error) {
					s, ok := from.(string)
					if !ok {
						return nil, errors.New("src type not matching to hin.HID")
					}
					return ParseID(s), nil
				},
			},
			{
				SrcType: time.Time{},
				DstType: _Int64,
				Fn: func(from any) (any, error) {
					s, ok := from.(time.Time)
					if !ok {
						return nil, errors.New("src type not matching to hin.HID")
					}
					return s.UnixMilli(), nil
				},
			},
			{
				SrcType: _Int64,
				DstType: time.Time{},
				Fn: func(from any) (any, error) {
					s, ok := from.(int64)
					if !ok {
						return nil, errors.New("src type not matching to hin.HID")
					}
					return time.UnixMilli(s), nil
				},
			},
		},
	}
	return copier.CopyWithOption(to, from, opts)
}
