package repository

import (
	"context"
	"errors"
	"member-shop-api/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ShiftRepository interface {
	GetCurrentOpenShift(ctx context.Context) (*model.CashShift, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*model.CashShift, error)
	Create(ctx context.Context, shift *model.CashShift) error
	Close(ctx context.Context, id bson.ObjectID, actualCash int64, notes, nextStaff string) (*model.CashShift, error)
	RecordSale(ctx context.Context, amount int64, paymentMethod string) error
	GetHistory(ctx context.Context, limit, offset int64) ([]model.CashShift, int64, error)
}

type shiftRepo struct {
	coll *mongo.Collection
}

func NewShiftRepository(db *mongo.Database) ShiftRepository {
	return &shiftRepo{coll: db.Collection("cash_shifts")}
}

func (r *shiftRepo) GetCurrentOpenShift(ctx context.Context) (*model.CashShift, error) {
	var s model.CashShift
	err := r.coll.FindOne(ctx, bson.M{"status": model.ShiftStatusOpen}).Decode(&s)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *shiftRepo) GetByID(ctx context.Context, id bson.ObjectID) (*model.CashShift, error) {
	var s model.CashShift
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shiftRepo) Create(ctx context.Context, shift *model.CashShift) error {
	shift.OpenedAt = time.Now()
	shift.Status = model.ShiftStatusOpen
	shift.ExpectedCash = shift.InitialFloat
	res, err := r.coll.InsertOne(ctx, shift)
	if err != nil {
		return err
	}
	shift.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *shiftRepo) Close(ctx context.Context, id bson.ObjectID, actualCash int64, notes, nextStaff string) (*model.CashShift, error) {
	shift, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shift.Status == model.ShiftStatusClosed {
		return nil, errors.New("ca làm việc này đã được chốt két trước đó")
	}

	now := time.Now()
	diff := actualCash - shift.ExpectedCash

	updates := bson.M{
		"actualCash":    actualCash,
		"difference":    diff,
		"closedAt":      now,
		"status":        model.ShiftStatusClosed,
		"notes":         notes,
		"nextStaffName": nextStaff,
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var closed model.CashShift
	err = r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&closed)
	if err != nil {
		return nil, err
	}
	return &closed, nil
}

func (r *shiftRepo) RecordSale(ctx context.Context, amount int64, paymentMethod string) error {
	current, err := r.GetCurrentOpenShift(ctx)
	if err != nil || current == nil {
		return nil // không có ca mở thì bỏ qua
	}

	incFields := bson.M{
		"totalSales": amount,
		"orderCount": 1,
	}
	if paymentMethod == "cash" {
		incFields["cashSales"] = amount
		incFields["expectedCash"] = amount
	} else {
		incFields["qrSales"] = amount
	}

	_, err = r.coll.UpdateOne(ctx, bson.M{"_id": current.ID}, bson.M{"$inc": incFields})
	return err
}

func (r *shiftRepo) GetHistory(ctx context.Context, limit, offset int64) ([]model.CashShift, int64, error) {
	total, err := r.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "openedAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if offset > 0 {
		opts.SetSkip(offset)
	}
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var list []model.CashShift
	if err := cursor.All(ctx, &list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
