package hin

import (
	"context"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type BaseService[E any] interface {
	Remove(ctx context.Context, query any) error
	Find(ctx context.Context, query any) ([]E, error)
	FindOne(ctx context.Context, query any) (E, error)
	Paging(ctx context.Context, query any, paging PagingQuery) (PagingDTO, error)
}

type BaseSrv[E any] struct {
	Logger *Logger
	Repo   BaseRepository[E]
}

func NewBaseService[E any](
	logger *Logger,
	repo BaseRepository[E],
) *BaseSrv[E] {
	return &BaseSrv[E]{
		logger,
		repo,
	}
}

func (s *BaseSrv[E]) FindOne(ctx context.Context, query any) (E, error) {
	d, r := s.Repo.FindOne(ctx, Criteria(query))
	return d, r.Error
}

func (s *BaseSrv[E]) Find(ctx context.Context, query any) ([]E, error) {
	d, r := s.Repo.Find(ctx, Criteria(query))
	return d, r.Error
}

func (s *BaseSrv[E]) Paging(ctx context.Context, query any, paging PagingQuery) (PagingDTO, error) {
	dto := PagingDTO{
		Page:  paging.Page,
		Count: paging.Count,
	}

	items, total, r := s.Repo.Paging(ctx, Criteria(query), paging)
	if r.Error != nil {
		s.Logger.Error("baseSrv.Paging", zap.Error(r.Error))
		return dto, r.Error
	}

	dto.Items = items
	dto.Total = total

	return dto, nil
}

func (s *BaseSrv[E]) Remove(ctx context.Context, query any) error {
	return s.Repo.Remove(ctx, Criteria(query)).Error
}

type BaseConverter[M any, E any] interface {
	ToModel(e E) M
	ToEntity(m M) E
	ToEntities(ms []M) []E
	ToModels(es []E) []M
}

type BaseDAO[T any] interface {
	Insert(ctx context.Context, model T) *MDR
	InsertMany(ctx context.Context, model []T) *MDR
	Find(ctx context.Context, filter any) ([]T, *MDR)
	FindOne(ctx context.Context, filter any) (T, *MDR)
	Update(ctx context.Context, filter any, model any) *MDR
	UpdateById(ctx context.Context, id any, model any) *MDR
	UpdateMany(ctx context.Context, filter any, model []any) *MDR
	CreateIndexes(ctx context.Context, models []mongo.IndexModel) *MDR
	Paging(ctx context.Context, filter any, paging PagingQuery) ([]T, int64, *MDR)
}

type BaseRepository[E any] interface {
	Save(ctx context.Context, entity E) *MDR
	Exist(ctx context.Context, filter CriteriaBuilder) bool
	Remove(ctx context.Context, filter CriteriaBuilder) *MDR
	Find(ctx context.Context, filter CriteriaBuilder) ([]E, *MDR)
	FindOne(ctx context.Context, filter CriteriaBuilder) (E, *MDR)
	Paging(ctx context.Context, filter CriteriaBuilder, paging PagingQuery) ([]E, int64, *MDR)
}

// MDR is Mongo Database Result
type MDR struct {
	Count int64
	Error error
	IDs   []string
}

func (m *MDR) ID() string {
	return m.IDs[0]
}

func (m *MDR) setID(id any) *MDR {
	if id == nil {
		return m
	}

	if v, ok := id.([]any); ok {
		m.Count = int64(len(v))
		for _, i := range v {
			m.IDs = append(m.IDs, i.(string))
		}
	} else {
		m.Count = 1
		m.IDs = append(m.IDs, id.(string))
	}
	return m
}

func (m *MDR) SetCount(count int64) *MDR {
	m.Count = count
	return m
}

func newErrMDR(err error) *MDR {
	return &MDR{Error: err}
}

type BaseRepo[M any, E any] struct {
	Dao           BaseDAO[M]
	Logger        *Logger
	Cv            BaseConverter[M, E]
	TypeConverter []copier.TypeConverter
}

func NewBaseRepository[M any, E any](
	dao BaseDAO[M],
	logger *Logger,
) *BaseRepo[M, E] {
	return &BaseRepo[M, E]{
		dao,
		logger,
		nil,
		make([]copier.TypeConverter, 0),
	}
}

func (r *BaseRepo[M, E]) WithModelConverter(cv BaseConverter[M, E]) *BaseRepo[M, E] {
	r.Cv = cv
	return r
}

func (r *BaseRepo[M, E]) WithTypeConverter(tc []copier.TypeConverter) *BaseRepo[M, E] {
	r.TypeConverter = tc
	return r
}

func (r *BaseRepo[M, E]) ToEntities(ms []M) []E {
	if r.Cv != nil {
		return r.Cv.ToEntities(ms)
	}

	es := make([]E, 0)
	for _, m := range ms {
		es = append(es, r.ToEntity(m))
	}
	return es
}

func (r *BaseRepo[M, E]) ToModels(es []E) []M {
	if r.Cv != nil {
		return r.Cv.ToModels(es)
	}

	ms := make([]M, 0)
	for _, e := range es {
		ms = append(ms, r.ToModel(e))
	}
	return ms
}

func (r *BaseRepo[M, E]) ToModel(e E) M {
	if r.Cv != nil {
		return r.Cv.ToModel(e)
	}

	var m M
	if err := Copy(&m, e, WithCopyConverters(r.TypeConverter)); err != nil {
		r.Logger.Error("Error Copier Entity toModel", zap.Error(err))
	}

	return m
}

func (r *BaseRepo[M, E]) ToEntity(m M) E {
	if r.Cv != nil {
		return r.Cv.ToEntity(m)
	}

	var e E
	if err := Copy(&e, m, WithCopyConverters(r.TypeConverter)); err != nil {
		r.Logger.Error("Error Copier Model toEntity", zap.Error(err))
	}

	return e
}

func (r *BaseRepo[M, E]) Save(ctx context.Context, entity E) *MDR {
	m := r.ToModel(entity)
	rv := reflect.ValueOf(&m)
	if v := rv.Elem().FieldByName("ID"); v.String() != "00000000000000000000" {
		// before update data set updated_at
		if v := rv.Elem().FieldByName("UpdatedAt"); v.IsValid() {
			v.Set(reflect.ValueOf(time.Now()))
		}
		return r.Dao.UpdateById(ctx, v.String(), m)
	} else {
		v.SetString(NewID().String())
		return r.Dao.Insert(ctx, m)
	}
}

func (r *BaseRepo[M, E]) Find(ctx context.Context, filter CriteriaBuilder) ([]E, *MDR) {
	if ms, dr := r.Dao.Find(ctx, filter.Mgo()); dr.Error != nil {
		return nil, dr
	} else {
		return r.ToEntities(ms), dr
	}
}

func (r *BaseRepo[M, E]) FindOne(ctx context.Context, filter CriteriaBuilder) (E, *MDR) {
	if m, dr := r.Dao.FindOne(ctx, filter.Mgo()); dr.Error != nil {
		var e E
		return e, dr
	} else {
		return r.ToEntity(m), dr
	}
}

func (r *BaseRepo[M, E]) Paging(ctx context.Context, filter CriteriaBuilder, paging PagingQuery) ([]E, int64, *MDR) {
	if ms, count, dr := r.Dao.Paging(ctx, filter.Mgo(), paging); dr.Error != nil {
		return nil, 0, dr
	} else {
		return r.ToEntities(ms), count, dr
	}
}

func (r *BaseRepo[M, E]) Remove(ctx context.Context, filter CriteriaBuilder) *MDR {
	return r.Dao.Update(ctx, filter.Mgo(), bson.M{"deleted_at": time.Now()})
}

func (r *BaseRepo[M, E]) Exist(ctx context.Context, filter CriteriaBuilder) bool {
	_, mdr := r.FindOne(ctx, filter)
	return mdr.Count > 0
}

type BaseMongoDAO[T any] struct {
	Logger *Logger
	Client *mongo.Client
	Col    *mongo.Collection
	Db     *mongo.Database
}

type MongoDAOOptions struct {
	DB    string
	Table string
}

func NewMongoDAO[T any](
	logger *Logger,
	client *mongo.Client,
	opts *MongoDAOOptions,
) *BaseMongoDAO[T] {
	defaultDatabase := viper.GetString("mongo.database")
	if defaultDatabase == "" {
		logger.Warn("NewMongoDAO: defaultDatabase not set")
	}
	if opts.DB == "" {
		opts.DB = defaultDatabase
	}
	db := client.Database(opts.DB)
	col := db.Collection(opts.Table)
	return &BaseMongoDAO[T]{
		logger,
		client,
		col,
		db,
	}
}

func (d *BaseMongoDAO[T]) Insert(ctx context.Context, model T) *MDR {
	r, err := d.Col.InsertOne(ctx, model)
	return (&MDR{Error: err}).setID(r.InsertedID)
}

func (d *BaseMongoDAO[T]) InsertMany(ctx context.Context, model []T) *MDR {
	var ms []any
	for _, m := range model {
		ms = append(ms, m)
	}
	r, err := d.Col.InsertMany(ctx, ms)
	return (&MDR{Error: err}).setID(r.InsertedIDs)
}

func (d *BaseMongoDAO[T]) Update(ctx context.Context, filter any, model any) *MDR {
	r, err := d.Col.UpdateOne(ctx, filter, bson.M{"$set": model})
	return (&MDR{Error: err, Count: r.ModifiedCount}).setID(r.UpsertedID)
}

func (d *BaseMongoDAO[T]) UpdateById(ctx context.Context, id any, model any) *MDR {
	r, err := d.Col.UpdateByID(ctx, id, bson.M{"$set": model})
	return (&MDR{Error: err, Count: r.ModifiedCount}).setID(id)
}

func (d *BaseMongoDAO[T]) UpdateMany(ctx context.Context, filter any, model []any) *MDR {
	r, err := d.Col.UpdateMany(ctx, filter, bson.M{"$set": model})
	return (&MDR{Error: err, Count: r.ModifiedCount}).setID(r.UpsertedID)
}

func (d *BaseMongoDAO[T]) Find(ctx context.Context, filter any) ([]T, *MDR) {
	opts := new(mopt.FindOptions)
	opts.SetSort(bson.D{{"created_at", -1}})
	cur, err := d.Col.Find(ctx, filter, opts)
	defer cur.Close(ctx)
	if err != nil {
		return nil, newErrMDR(err)
	}

	r := make([]T, 0)
	for cur.Next(ctx) {
		var result T
		if err := cur.Decode(&result); err != nil {
			return nil, newErrMDR(err)
		}
		r = append(r, result)
	}

	return r, new(MDR).SetCount(int64(len(r)))
}

func (d *BaseMongoDAO[T]) FindOne(ctx context.Context, filter any) (T, *MDR) {
	opts := new(mopt.FindOneOptions)
	opts.SetSort(bson.D{{"created_at", -1}})
	cur := d.Col.FindOne(ctx, filter, opts)
	var r T
	if err := cur.Decode(&r); err != nil {
		return r, newErrMDR(err)
	}

	return r, new(MDR).SetCount(1)
}

func (d *BaseMongoDAO[T]) Paging(ctx context.Context, filter any, paging PagingQuery) ([]T, int64, *MDR) {
	opts := new(mopt.FindOptions)
	opts.SetSort(bson.D{{"created_at", -1}})
	opts.SetLimit(paging.Count)
	opts.SetSkip(paging.Count * paging.Page)

	cur, err := d.Col.Find(ctx, filter, opts)
	defer cur.Close(ctx)
	if err != nil {
		return nil, 0, newErrMDR(err)
	}

	r := make([]T, 0)
	for cur.Next(ctx) {
		var result T
		if err := cur.Decode(&result); err != nil {
			return nil, 0, newErrMDR(err)
		}
		r = append(r, result)
	}

	total, err := d.Col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, total, newErrMDR(err)
	}
	return r, total, new(MDR).SetCount(int64(len(r)))
}

func (d *BaseMongoDAO[T]) CreateIndexes(ctx context.Context, models []mongo.IndexModel) *MDR {
	_, err := d.Col.Indexes().CreateMany(ctx, models)
	return newErrMDR(err)
}
