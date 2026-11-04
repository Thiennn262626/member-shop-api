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

type StaffFilter struct {
	Search string
	Role   string
	Limit  int64
	Offset int64
}

type StaffRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.StaffAccount, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.StaffAccount, error)
	GetAll(ctx context.Context) ([]model.StaffAccount, error)
	GetList(ctx context.Context, filter StaffFilter) ([]model.StaffAccount, int64, error)
	Create(ctx context.Context, s *model.StaffAccount) error
	Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.StaffAccount, error)
	ToggleStatus(ctx context.Context, id bson.ObjectID) (*model.StaffAccount, error)
}

type staffRepo struct {
	coll *mongo.Collection
}

func NewStaffRepository(db *mongo.Database) StaffRepository {
	return &staffRepo{coll: db.Collection("staff_accounts")}
}

func (r *staffRepo) FindByUsername(ctx context.Context, username string) (*model.StaffAccount, error) {
	clean := strings.ToLower(strings.TrimSpace(username))
	var s model.StaffAccount
	err := r.coll.FindOne(ctx, bson.M{"username": clean}).Decode(&s)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *staffRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.StaffAccount, error) {
	var s model.StaffAccount
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *staffRepo) GetAll(ctx context.Context) ([]model.StaffAccount, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.StaffAccount
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *staffRepo) GetList(ctx context.Context, filter StaffFilter) ([]model.StaffAccount, int64, error) {
	cond := bson.M{}
	if filter.Role != "" && filter.Role != "ALL" {
		cond["role"] = filter.Role
	}
	if filter.Search != "" {
		cond["$or"] = []bson.M{
			{"name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"username": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"phone": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	total, err := r.coll.CountDocuments(ctx, cond)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
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

	var list []model.StaffAccount
	if err := cursor.All(ctx, &list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *staffRepo) Create(ctx context.Context, s *model.StaffAccount) error {
	s.Username = strings.ToLower(strings.TrimSpace(s.Username))
	s.CreatedAt = time.Now()
	res, err := r.coll.InsertOne(ctx, s)
	if err != nil {
		return err
	}
	s.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *staffRepo) Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.StaffAccount, error) {
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var updated model.StaffAccount
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *staffRepo) ToggleStatus(ctx context.Context, id bson.ObjectID) (*model.StaffAccount, error) {
	s, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.Username == "admin" {
		return nil, errors.New("không thể khóa tài khoản admin chính")
	}
	return r.Update(ctx, id, bson.M{"isActive": !s.IsActive})
}
