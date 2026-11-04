package repository

import (
	"context"
	"errors"
	"member-shop-api/internal/model"
	"member-shop-api/pkg/apperror"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrderFilter struct {
	Status        string
	PaymentStatus string
	Search        string
	CustomerPhone string
	StartDate     *time.Time
	EndDate       *time.Time
	Limit         int64
	Offset        int64
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id bson.ObjectID) (*model.Order, error)
	GetByCode(ctx context.Context, code string) (*model.Order, error)
	GetByPhone(ctx context.Context, phone string) ([]model.Order, error)
	GetList(ctx context.Context, filter OrderFilter) ([]model.Order, int64, error)
	Update(ctx context.Context, id bson.ObjectID, updates bson.M, expectedVersion ...int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, expectedVersion ...int64) (*model.Order, error)
}

type orderRepo struct {
	coll *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) OrderRepository {
	return &orderRepo{coll: db.Collection("orders")}
}

func (r *orderRepo) Create(ctx context.Context, order *model.Order) error {
	now := time.Now()
	order.CreatedAt = now
	res, err := r.coll.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	order.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *orderRepo) GetByID(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	var o model.Order
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) GetByCode(ctx context.Context, code string) (*model.Order, error) {
	clean := strings.TrimSpace(code)
	if !strings.HasPrefix(clean, "#") {
		clean = "#" + clean
	}
	var o model.Order
	err := r.coll.FindOne(ctx, bson.M{"orderCode": clean}).Decode(&o)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) GetByPhone(ctx context.Context, phone string) ([]model.Order, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"customerPhone": phone}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepo) GetList(ctx context.Context, filter OrderFilter) ([]model.Order, int64, error) {
	query := bson.M{}

	if filter.Status != "" && filter.Status != "ALL" {
		query["status"] = filter.Status
	}
	if filter.PaymentStatus != "" {
		query["paymentStatus"] = filter.PaymentStatus
	}
	if filter.CustomerPhone != "" {
		query["customerPhone"] = filter.CustomerPhone
	}
	if filter.Search != "" {
		s := strings.TrimSpace(filter.Search)
		query["$or"] = []bson.M{
			{"orderCode": bson.M{"$regex": s, "$options": "i"}},
			{"customerPhone": bson.M{"$regex": s, "$options": "i"}},
			{"customerName": bson.M{"$regex": s, "$options": "i"}},
		}
	}
	if filter.StartDate != nil || filter.EndDate != nil {
		timeQuery := bson.M{}
		if filter.StartDate != nil {
			timeQuery["$gte"] = *filter.StartDate
		}
		if filter.EndDate != nil {
			timeQuery["$lte"] = *filter.EndDate
		}
		query["createdAt"] = timeQuery
	}

	total, err := r.coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if filter.Limit > 0 {
		findOpts.SetLimit(filter.Limit)
	}
	if filter.Offset > 0 {
		findOpts.SetSkip(filter.Offset)
	}

	cursor, err := r.coll.Find(ctx, query, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepo) Update(ctx context.Context, id bson.ObjectID, updates bson.M, expectedVersion ...int64) (*model.Order, error) {
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	filter := bson.M{"_id": id}
	if len(expectedVersion) > 0 && expectedVersion[0] >= 0 {
		expVer := expectedVersion[0]
		filter["$or"] = []bson.M{
			{"version": expVer},
			{"version": bson.M{"$exists": false}},
		}
	}

	delete(updates, "version")
	delete(updates, "_id")
	delete(updates, "id")

	updateDoc := bson.M{
		"$set": updates,
		"$inc": bson.M{"version": 1},
	}

	var updated model.Order
	err := r.coll.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updated)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			count, checkErr := r.coll.CountDocuments(ctx, bson.M{"_id": id})
			if checkErr == nil && count > 0 {
				return nil, apperror.VersionMismatch("Đơn hàng đã được cập nhật bởi thao tác khác. Vui lòng tải lại dữ liệu mới nhất.")
			}
			return nil, apperror.NotFound("Không tìm thấy đơn hàng yêu cầu")
		}
		return nil, err
	}
	return &updated, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, status model.OrderStatus, expectedVersion ...int64) (*model.Order, error) {
	now := time.Now()
	updates := bson.M{"status": status}

	switch status {
	case model.OrderStatusConfirmed:
		updates["confirmedAt"] = now
	case model.OrderStatusShipping:
		updates["shippingAt"] = now
	case model.OrderStatusCompleted:
		updates["completedAt"] = now
	case model.OrderStatusCancelled:
		updates["cancelledAt"] = now
	}

	return r.Update(ctx, id, updates, expectedVersion...)
}
