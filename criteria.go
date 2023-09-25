package hin

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

type CriteriaBuilder struct {
	Query any
	SQL   string
	Vars  []any
	Error error
}

func Criteria(query any, args ...any) CriteriaBuilder {
	builder := CriteriaBuilder{
		Query: query,
		Vars:  args,
	}

	if s, ok := query.(string); ok {
		builder.SQL = s

		if strings.Count(s, "?") != len(args) {
			builder.Error = errors.New("the number of args is not equal to given query")
		}
	}

	return builder
}

func (c *CriteriaBuilder) Mgo() bson.M {
	if c.Error != nil {
		return nil
	}

	if c.SQL != "" {
		return buildMgoSql(c.SQL, c.Vars...)
	}

	return buildMgoEntity(c.Query)
}

func buildMgoSql(sql string, args ...any) bson.M {
	bm := bson.M{
		"deleted_at": nil,
	}

	vc := 0
	for _, sep := range strings.Split(sql, " AND ") {
		q := strings.Split(sep, " ")
		k := q[0]
		var v any

		if k == "id" && !strings.Contains(sql, "_id") {
			k = "_id"
		}

		if "?" == q[2] {
			v = args[vc]
			vc++
		} else {
			v = q[2]
		}

		switch q[1] {
		case "=":
			bm[k] = v
		case "!=":
			bm[k] = bson.M{"$ne": v}
		case "like":
			bm[k] = bson.M{"$regex": v}
		case "in":
			bm[k] = bson.M{"$in": v}
		}
	}
	return bm
}

func buildMgoEntity(entity any) bson.M {
	bm := bson.M{
		"deleted_at": nil,
	}

	if entity == nil || entity == "" {
		return bm
	}

	v := reflect.ValueOf(entity)
	for i := 0; i < v.NumField(); i++ {

		field := v.Type().Field(i)
		tag := field.Tag

		label := tag.Get("criteria")
		value := v.Field(i)
		key := toSnake(field.Name)

		if label == "-" || !value.IsValid() || value.IsZero() {
			if strings.HasPrefix(label, "nil") {
				bm[key] = nil
			}
			if strings.HasPrefix(label, "empty") {
				bm[key] = value.Interface()
			}
			continue
		}

		if label == "" {
			bm[key] = value.Interface()
			continue
		}

		lv := strings.Split(label, ",")
		if len(lv) > 1 {
			key = lv[1]
		}

		switch lv[0] {
		case "=":
		case "eq":
			bm[key] = value.Interface()
		case "!=":
		case "ne":
			bm[key] = bson.M{"$ne": value.Interface()}
		case "like":
			bm[key] = bson.M{"$regex": value.Interface()}
		case "in":
			bm[key] = bson.M{"$in": value.Interface()}
		default:
			bm[key] = value.Interface()
		}
	}

	special := map[string]string{
		"id": "_id",
	}
skl:
	for sk, sv := range special {
		for k, _ := range bm {
			if k == sv {
				continue skl
			}
		}
		if v := bm[sk]; v != nil {
			bm[sv] = bm[sk]
			delete(bm, sk)
		}
	}

	return bm
}

func toSnake(s string) string {
	if strings.ToUpper(s) == s {
		return strings.ToLower(s)
	}
	var ret []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			ret = append(ret, '_')
		}
		ret = append(ret, r)
	}
	return strings.ToLower(string(ret))
}
