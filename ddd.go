package hin

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	return s.Repo.FindOne(ctx, Criteria(query))
}

func (s *BaseSrv[E]) Find(ctx context.Context, query any) ([]E, error) {
	return s.Repo.Find(ctx, Criteria(query))
}

func (s *BaseSrv[E]) Paging(ctx context.Context, query any, paging PagingQuery) (PagingDTO, error) {
	dto := PagingDTO{
		Page:  paging.Page,
		Count: paging.Count,
	}

	items, total, err := s.Repo.Paging(ctx, Criteria(query), paging)
	if err != nil {
		s.Logger.Error("baseSrv.Paging", zap.Error(err))
		return dto, err
	}

	dto.Items = items
	dto.Total = total

	return dto, nil
}

func (s *BaseSrv[E]) Remove(ctx context.Context, query any) error {
	return s.Repo.Remove(ctx, Criteria(query))
}

type BaseConverter[M any, E any] interface {
	ToModel(e E) M
	ToEntity(m M) E
}

type BaseDAO[T any] interface {
	Insert(ctx context.Context, model T) error
	InsertMany(ctx context.Context, model []T) error
	Find(ctx context.Context, filter any) ([]T, error)
	FindOne(ctx context.Context, filter any) (T, error)
	Update(ctx context.Context, filter any, model any) error
	UpdateById(ctx context.Context, id any, model any) error
	UpdateMany(ctx context.Context, filter any, model []any) error
	CreateIndexes(ctx context.Context, models []mongo.IndexModel) error
	Paging(ctx context.Context, filter any, paging PagingQuery) ([]T, int64, error)
}

type BaseRepository[E any] interface {
	Save(ctx context.Context, entity E) error
	Exist(ctx context.Context, filter CriteriaBuilder) bool
	Remove(ctx context.Context, filter CriteriaBuilder) error
	Find(ctx context.Context, filter CriteriaBuilder) ([]E, error)
	FindOne(ctx context.Context, filter CriteriaBuilder) (E, error)
	Paging(ctx context.Context, filter CriteriaBuilder, paging PagingQuery) ([]E, int64, error)
}

type BaseRepo[M any, E any] struct {
	Dao    BaseDAO[M]
	Cv     BaseConverter[M, E]
	Logger *Logger
}

func NewBaseRepository[M any, E any](
	dao BaseDAO[M],
	cv BaseConverter[M, E],
	logger *Logger,
) *BaseRepo[M, E] {
	return &BaseRepo[M, E]{
		dao,
		cv,
		logger,
	}
}

func (r *BaseRepo[M, E]) toEntities(ms []M) []E {
	es := make([]E, 0)
	for _, m := range ms {
		es = append(es, r.Cv.ToEntity(m))
	}
	return es
}

func (r *BaseRepo[M, E]) Save(ctx context.Context, entity E) error {
	m := r.Cv.ToModel(entity)
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

func (r *BaseRepo[M, E]) Find(ctx context.Context, filter CriteriaBuilder) ([]E, error) {
	if ms, err := r.Dao.Find(ctx, filter.Mgo()); err != nil {
		return nil, err
	} else {
		return r.toEntities(ms), nil
	}
}

func (r *BaseRepo[M, E]) FindOne(ctx context.Context, filter CriteriaBuilder) (E, error) {
	if m, err := r.Dao.FindOne(ctx, filter.Mgo()); err != nil {
		var e E
		return e, err
	} else {
		return r.Cv.ToEntity(m), nil
	}
}

func (r *BaseRepo[M, E]) Paging(ctx context.Context, filter CriteriaBuilder, paging PagingQuery) ([]E, int64, error) {
	if ms, count, err := r.Dao.Paging(ctx, filter.Mgo(), paging); err != nil {
		return nil, 0, err
	} else {
		return r.toEntities(ms), count, nil
	}
}

func (r *BaseRepo[M, E]) Remove(ctx context.Context, filter CriteriaBuilder) error {
	return r.Dao.Update(ctx, filter.Mgo(), bson.M{"deleted_at": time.Now()})
}

func (r *BaseRepo[M, E]) Exist(ctx context.Context, filter CriteriaBuilder) bool {
	var e E
	entity, err := r.FindOne(ctx, filter)
	return err == nil && !reflect.DeepEqual(entity, e)
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
	db := client.Database(opts.DB)
	col := db.Collection(opts.Table)
	return &BaseMongoDAO[T]{
		logger,
		client,
		col,
		db,
	}
}

func (d *BaseMongoDAO[T]) Insert(ctx context.Context, model T) error {
	_, err := d.Col.InsertOne(ctx, model)
	return err
}

func (d *BaseMongoDAO[T]) InsertMany(ctx context.Context, model []T) error {
	var ms []any
	for _, m := range model {
		ms = append(ms, m)
	}
	_, err := d.Col.InsertMany(ctx, ms)
	return err
}

func (d *BaseMongoDAO[T]) Update(ctx context.Context, filter any, model any) error {
	_, err := d.Col.UpdateOne(ctx, filter, bson.M{"$set": model})
	return err
}

func (d *BaseMongoDAO[T]) UpdateById(ctx context.Context, id any, model any) error {
	_, err := d.Col.UpdateByID(ctx, id, bson.M{"$set": model})
	return err
}

func (d *BaseMongoDAO[T]) UpdateMany(ctx context.Context, filter any, model []any) error {
	_, err := d.Col.UpdateMany(ctx, filter, bson.M{"$set": model})
	return err
}

func (d *BaseMongoDAO[T]) Find(ctx context.Context, filter any) ([]T, error) {
	opts := new(options.FindOptions)
	opts.SetSort(bson.D{{"created_at", -1}})
	cur, err := d.Col.Find(ctx, filter, opts)
	defer cur.Close(ctx)
	if err != nil {
		return nil, err
	}

	r := make([]T, 0)
	for cur.Next(ctx) {
		var result T
		if err := cur.Decode(&result); err != nil {
			return nil, err
		}
		r = append(r, result)
	}

	return r, nil
}

func (d *BaseMongoDAO[T]) FindOne(ctx context.Context, filter any) (T, error) {
	opts := new(options.FindOneOptions)
	opts.SetSort(bson.D{{"created_at", -1}})
	cur := d.Col.FindOne(ctx, filter, opts)
	var r T
	if err := cur.Decode(&r); err != nil {
		return r, err
	}
	return r, nil
}

func (d *BaseMongoDAO[T]) Paging(ctx context.Context, filter any, paging PagingQuery) ([]T, int64, error) {
	opts := new(options.FindOptions)
	opts.SetSort(bson.D{{"created_at", -1}})
	opts.SetLimit(paging.Count)
	opts.SetSkip(paging.Count * paging.Page)

	cur, err := d.Col.Find(ctx, filter, opts)
	defer cur.Close(ctx)
	if err != nil {
		return nil, 0, err
	}

	r := make([]T, 0)
	for cur.Next(ctx) {
		var result T
		if err := cur.Decode(&result); err != nil {
			return nil, 0, err
		}
		r = append(r, result)
	}

	total, err := d.Col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, total, err
	}
	return r, total, nil
}

func (d *BaseMongoDAO[T]) CreateIndexes(ctx context.Context, models []mongo.IndexModel) error {
	_, err := d.Col.Indexes().CreateMany(ctx, models)
	return err
}
