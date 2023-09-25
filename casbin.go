package hin

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

var casbinDefaultConf = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`

func NewCasbin(mongoCli *mongo.Client) *casbin.Enforcer {
	adapter, err := mongodbadapter.NewAdapterByDB(mongoCli, &mongodbadapter.AdapterConfig{
		DatabaseName:   viper.GetString("casbin.database"),
		CollectionName: viper.GetString("casbin.collection"),
	})
	if err != nil {
		panic(err)
	}

	var m model.Model
	modelFile := viper.GetString("casbin.model")
	if modelFile == "" {
		m, _ = model.NewModelFromString(casbinDefaultConf)
	} else {
		m, err = model.NewModelFromFile(modelFile)
		if err != nil {
			panic(err)
		}
	}

	c, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		panic(err)
	}

	return c
}
