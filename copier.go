package hin

import (
	"errors"
	"github.com/jinzhu/copier"
	"reflect"
	"time"
)

var (
	_Int64 int64 = 0
)

type CopyOption func(co *copier.Option)

func WithCopyConverters(opts []copier.TypeConverter) CopyOption {
	return func(co *copier.Option) {
		co.Converters = append(co.Converters, opts...)
	}
}

func WithCopyIgnoreEmpty(v bool) CopyOption {
	return func(co *copier.Option) {
		co.IgnoreEmpty = v
	}
}

func WithCopyCaseSensitive(v bool) CopyOption {
	return func(co *copier.Option) {
		co.CaseSensitive = v
	}
}

func WithCopyDeep(v bool) CopyOption {
	return func(co *copier.Option) {
		co.DeepCopy = v
	}
}

func WithFilter(from any, to any, skipOrOnly bool, keys ...string) CopyOption {
	return func(co *copier.Option) {
		fields := make([]string, 0)
		mapping := map[string]string{}
		t := reflect.TypeOf(from)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fields = append(fields, field.Name)
		}

		for _, k := range fields {

			if skipOrOnly {
				mapping[k] = k
			} else {
				mapping[k] = "_"
			}

			for _, key := range keys {
				if k == key {
					if !skipOrOnly {
						mapping[k] = k
					} else {
						mapping[k] = "_"
					}
				}
			}
		}

		co.FieldNameMapping = []copier.FieldNameMapping{
			{SrcType: from, DstType: to, Mapping: mapping},
		}
	}
}

func Copy(to any, from any, opts ...CopyOption) error {

	copierOption := copier.Option{
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

	for _, o := range opts {
		o(&copierOption)
	}

	return copier.CopyWithOption(to, from, copierOption)
}
