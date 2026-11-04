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

type CustomerRepository interface {
	FindByPhone(ctx context.Context, phone string) (*model.Customer, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.Customer, error)
	GetAll(ctx context.Context) ([]model.Customer, error)
	CreateOrGet(ctx context.Context, name, phone string) (*model.Customer, error)
	Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.Customer, error)
	AdjustPoints(ctx context.Context, id bson.ObjectID, delta int64) (*model.Customer, error)
	UpdateStatsAfterOrder(ctx context.Context, id bson.ObjectID, totalSpentDelta, pointsEarned, pointsRedeemed int64) (*model.Customer, error)
	DeductPointsAtomic(ctx context.Context, id bson.ObjectID, pointsToDeduct int64) (*model.Customer, error)
	RefundPointsAtomic(ctx context.Context, id bson.ObjectID, pointsToRefund int64) (*model.Customer, error)
	AddSpentAndEarnedPoints(ctx context.Context, id bson.ObjectID, totalSpentDelta, pointsEarned int64) (*model.Customer, error)
}

type customerRepo struct {
	coll *mongo.Collection
}

func NewCustomerRepository(db *mongo.Database) CustomerRepository {
	return &customerRepo{coll: db.Collection("customers")}
}

func (r *customerRepo) FindByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	var customer model.Customer
	err := r.coll.FindOne(ctx, bson.M{"phone": phone}).Decode(&customer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.Customer, error) {
	var customer model.Customer
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) GetAll(ctx context.Context) ([]model.Customer, error) {
	opts := options.Find().SetSort(bson.D{{Key: "totalSpent", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.Customer
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *customerRepo) CreateOrGet(ctx context.Context, name, phone string) (*model.Customer, error) {
	existing, err := r.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if name != "" && existing.Name != name {
			return r.Update(ctx, existing.ID, bson.M{"name": name})
		}
		return existing, nil
	}

	now := time.Now()
	newCust := model.Customer{
		Phone:           phone,
		Name:            name,
		Tier:            model.TierMember,
		EarnedPoints:    0,
		RedeemedPoints:  0,
		AvailablePoints: 0,
		TotalSpent:      0,
		JoinedDate:      now.Format("02/01/2006"),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if newCust.Name == "" {
		if len(phone) >= 4 {
			newCust.Name = "Khách " + phone[len(phone)-4:]
		} else {
			newCust.Name = "Khách Hàng"
		}
	}

	res, err := r.coll.InsertOne(ctx, newCust)
	if err != nil {
		return nil, err
	}
	newCust.ID = res.InsertedID.(bson.ObjectID)
	return &newCust, nil
}

func (r *customerRepo) Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.Customer, error) {
	updates["updatedAt"] = time.Now()
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var cust model.Customer
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&cust)
	if err != nil {
		return nil, err
	}
	return &cust, nil
}

func (r *customerRepo) AdjustPoints(ctx context.Context, id bson.ObjectID, delta int64) (*model.Customer, error) {
	cust, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newAvailable := cust.AvailablePoints + delta
	if newAvailable < 0 {
		newAvailable = 0
	}

	updates := bson.M{
		"availablePoints": newAvailable,
		"updatedAt":       time.Now(),
	}
	if delta > 0 {
		updates["earnedPoints"] = cust.EarnedPoints + delta
	} else {
		updates["redeemedPoints"] = cust.RedeemedPoints + (-delta)
	}

	return r.Update(ctx, id, updates)
}

func (r *customerRepo) UpdateStatsAfterOrder(ctx context.Context, id bson.ObjectID, totalSpentDelta, pointsEarned, pointsRedeemed int64) (*model.Customer, error) {
	cust, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newTotalSpent := cust.TotalSpent + totalSpentDelta
	newTier := model.CalculateTier(newTotalSpent)
	newEarned := cust.EarnedPoints + pointsEarned
	newRedeemed := cust.RedeemedPoints + pointsRedeemed
	newAvailable := newEarned - newRedeemed
	if newAvailable < 0 {
		newAvailable = 0
	}

	updates := bson.M{
		"totalSpent":      newTotalSpent,
		"tier":            newTier,
		"earnedPoints":    newEarned,
		"redeemedPoints":  newRedeemed,
		"availablePoints": newAvailable,
		"updatedAt":       time.Now(),
	}
	return r.Update(ctx, id, updates)
}

// DeductPointsAtomic trừ điểm thưởng ngay lập tức với điều kiện Atomic chống Double-Spending (SEC-06)
func (r *customerRepo) DeductPointsAtomic(ctx context.Context, id bson.ObjectID, pointsToDeduct int64) (*model.Customer, error) {
	if pointsToDeduct <= 0 {
		return r.FindByID(ctx, id)
	}

	filter := bson.M{
		"_id":             id,
		"availablePoints": bson.M{"$gte": pointsToDeduct},
	}
	updates := bson.M{
		"$inc": bson.M{
			"availablePoints": -pointsToDeduct,
			"redeemedPoints":  pointsToDeduct,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var cust model.Customer
	err := r.coll.FindOneAndUpdate(ctx, filter, updates, opts).Decode(&cust)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("insufficient_points")
		}
		return nil, err
	}
	return &cust, nil
}

// RefundPointsAtomic hoàn trả điểm khi đơn hàng bị hủy (SEC-06)
func (r *customerRepo) RefundPointsAtomic(ctx context.Context, id bson.ObjectID, pointsToRefund int64) (*model.Customer, error) {
	if pointsToRefund <= 0 {
		return r.FindByID(ctx, id)
	}

	updates := bson.M{
		"$inc": bson.M{
			"availablePoints": pointsToRefund,
			"redeemedPoints":  -pointsToRefund,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var cust model.Customer
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, updates, opts).Decode(&cust)
	if err != nil {
		return nil, err
	}
	return &cust, nil
}

// AddSpentAndEarnedPoints cộng doanh số và điểm tích lũy khi đơn hàng chuyển sang Completed (SEC-06)
func (r *customerRepo) AddSpentAndEarnedPoints(ctx context.Context, id bson.ObjectID, totalSpentDelta, pointsEarned int64) (*model.Customer, error) {
	cust, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cust == nil {
		return nil, errors.New("customer not found")
	}

	newTotalSpent := cust.TotalSpent + totalSpentDelta
	newTier := model.CalculateTier(newTotalSpent)
	newEarned := cust.EarnedPoints + pointsEarned
	newAvailable := cust.AvailablePoints + pointsEarned

	updates := bson.M{
		"totalSpent":      newTotalSpent,
		"tier":            newTier,
		"earnedPoints":    newEarned,
		"availablePoints": newAvailable,
		"updatedAt":       time.Now(),
	}
	return r.Update(ctx, id, updates)
}
