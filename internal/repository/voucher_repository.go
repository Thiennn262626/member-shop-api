package repository

import (
	"context"
	"errors"
	"member-shop-api/internal/model"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type VoucherRepository interface {
	GetByCode(ctx context.Context, code string) (*model.Voucher, error)
	GetAll(ctx context.Context) ([]model.Voucher, error)
	Create(ctx context.Context, v *model.Voucher) error
	Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.Voucher, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	IncrementUsedCount(ctx context.Context, id bson.ObjectID) error
}

type voucherRepo struct {
	coll *mongo.Collection
}

func NewVoucherRepository(db *mongo.Database) VoucherRepository {
	return &voucherRepo{coll: db.Collection("vouchers")}
}

func (r *voucherRepo) GetByCode(ctx context.Context, code string) (*model.Voucher, error) {
	clean := strings.ToUpper(strings.TrimSpace(code))
	var v model.Voucher
	err := r.coll.FindOne(ctx, bson.M{"code": clean}).Decode(&v)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *voucherRepo) GetAll(ctx context.Context) ([]model.Voucher, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.Voucher
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *voucherRepo) Create(ctx context.Context, v *model.Voucher) error {
	v.Code = strings.ToUpper(strings.TrimSpace(v.Code))
	v.CreatedAt = time.Now()
	res, err := r.coll.InsertOne(ctx, v)
	if err != nil {
		return err
	}
	v.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *voucherRepo) Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.Voucher, error) {
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var updated model.Voucher
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *voucherRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *voucherRepo) IncrementUsedCount(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"usedCount": 1}})
	return err
}
