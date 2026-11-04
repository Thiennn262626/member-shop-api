package repository

import (
	"context"
	"member-shop-api/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductFilter struct {
	Category   string
	Search     string
	OnlyActive bool
	Limit      int64
	Offset     int64
}

type ProductRepository interface {
	GetAll(ctx context.Context, onlyActive bool) ([]model.Product, error)
	GetList(ctx context.Context, filter ProductFilter) ([]model.Product, int64, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*model.Product, error)
	Create(ctx context.Context, p *model.Product) error
	Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.Product, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	ToggleActive(ctx context.Context, id bson.ObjectID) (*model.Product, error)
	ToggleAllActive(ctx context.Context, isActive bool, category string) error
	ToggleOptionAvailability(ctx context.Context, optionID string, isAvailable bool) error
}

type productRepo struct {
	coll *mongo.Collection
}

func NewProductRepository(db *mongo.Database) ProductRepository {
	coll := db.Collection("products")
	return &productRepo{coll: coll}
}

func (r *productRepo) GetAll(ctx context.Context, onlyActive bool) ([]model.Product, error) {
	filter := bson.M{}
	if onlyActive {
		filter["isActive"] = true
	}
	findOpts := options.Find().SetSort(bson.D{{Key: "soldCount", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []model.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepo) GetList(ctx context.Context, filter ProductFilter) ([]model.Product, int64, error) {
	cond := bson.M{}
	if filter.OnlyActive {
		cond["isActive"] = true
	}
	if filter.Category != "" && filter.Category != "ALL" {
		cond["category"] = filter.Category
	}
	if filter.Search != "" {
		cond["$or"] = []bson.M{
			{"name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"description": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	total, err := r.coll.CountDocuments(ctx, cond)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "soldCount", Value: -1}, {Key: "createdAt", Value: -1}})
	if filter.Limit > 0 {
		opts.SetLimit(filter.Limit)
	}
	if filter.Offset > 0 {
		opts.SetSkip(filter.Offset)
	}

	cursor, err := r.coll.Find(ctx, cond, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var products []model.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepo) GetByID(ctx context.Context, id bson.ObjectID) (*model.Product, error) {
	var product model.Product
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) Create(ctx context.Context, p *model.Product) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	res, err := r.coll.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *productRepo) Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.Product, error) {
	updates["updatedAt"] = time.Now()
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var updated model.Product
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *productRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *productRepo) ToggleActive(ctx context.Context, id bson.ObjectID) (*model.Product, error) {
	prod, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.Update(ctx, id, bson.M{"isActive": !prod.IsActive})
}

func (r *productRepo) ToggleAllActive(ctx context.Context, isActive bool, category string) error {
	filter := bson.M{}
	if category != "" && category != "Tất cả" {
		filter["category"] = category
	}
	_, err := r.coll.UpdateMany(ctx, filter, bson.M{
		"$set": bson.M{
			"isActive":  isActive,
			"updatedAt": time.Now(),
		},
	})
	return err
}

func (r *productRepo) ToggleOptionAvailability(ctx context.Context, optionID string, isAvailable bool) error {
	filter := bson.M{"optionGroups.options.id": optionID}
	update := bson.M{
		"$set": bson.M{
			"optionGroups.$[].options.$[opt].isAvailable": isAvailable,
			"updatedAt": time.Now(),
		},
	}
	arrayFilters := options.UpdateMany().SetArrayFilters([]any{
		bson.M{"opt.id": optionID},
	})
	_, err := r.coll.UpdateMany(ctx, filter, update, arrayFilters)
	return err
}
