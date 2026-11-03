package repository

import (
	"context"
	"member-shop-api/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PointTransactionRepository interface {
	Create(ctx context.Context, tx *model.PointTransaction) error
	GetByCustomerID(ctx context.Context, customerID bson.ObjectID) ([]model.PointTransaction, error)
}

type pointTxRepo struct {
	coll *mongo.Collection
}

func NewPointTransactionRepository(db *mongo.Database) PointTransactionRepository {
	return &pointTxRepo{coll: db.Collection("point_transactions")}
}

func (r *pointTxRepo) Create(ctx context.Context, tx *model.PointTransaction) error {
	tx.CreatedAt = time.Now()
	res, err := r.coll.InsertOne(ctx, tx)
	if err != nil {
		return err
	}
	tx.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *pointTxRepo) GetByCustomerID(ctx context.Context, customerID bson.ObjectID) ([]model.PointTransaction, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"customerId": customerID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.PointTransaction
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
