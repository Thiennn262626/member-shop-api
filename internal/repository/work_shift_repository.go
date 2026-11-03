package repository

import (
	"context"
	"member-shop-api/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type WorkShiftRepository interface {
	GetByMonth(ctx context.Context, monthStr string, staffID *bson.ObjectID) ([]model.WorkShift, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*model.WorkShift, error)
	CreateBatch(ctx context.Context, shifts []model.WorkShift) error
	DeleteFutureScheduled(ctx context.Context, staffID bson.ObjectID, monthStr string) error
	Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.WorkShift, error)
	GetPayrollSummary(ctx context.Context, monthStr string) ([]model.MonthlyPayroll, error)
	SavePayrollAdjustment(ctx context.Context, monthStr string, staffID bson.ObjectID, allowance, deduction int64) error
	FinalizePayroll(ctx context.Context, monthStr string, finalizedBy string) error
}

type workShiftRepo struct {
	coll        *mongo.Collection
	payrollColl *mongo.Collection
	staffColl   *mongo.Collection
}

func NewWorkShiftRepository(db *mongo.Database) WorkShiftRepository {
	return &workShiftRepo{
		coll:        db.Collection("work_shifts"),
		payrollColl: db.Collection("monthly_payrolls"),
		staffColl:   db.Collection("staff_accounts"),
	}
}

func (r *workShiftRepo) GetByMonth(ctx context.Context, monthStr string, staffID *bson.ObjectID) ([]model.WorkShift, error) {
	filter := bson.M{
		"date": bson.M{"$regex": "^" + monthStr},
	}
	if staffID != nil {
		filter["staffId"] = *staffID
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}, {Key: "scheduledStartTime", Value: 1}})
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.WorkShift
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *workShiftRepo) GetByID(ctx context.Context, id bson.ObjectID) (*model.WorkShift, error) {
	var s model.WorkShift
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *workShiftRepo) CreateBatch(ctx context.Context, shifts []model.WorkShift) error {
	if len(shifts) == 0 {
		return nil
	}
	docs := make([]any, len(shifts))
	for i := range shifts {
		shifts[i].CreatedAt = time.Now()
		shifts[i].UpdatedAt = time.Now()
		docs[i] = shifts[i]
	}
	_, err := r.coll.InsertMany(ctx, docs)
	return err
}

func (r *workShiftRepo) DeleteFutureScheduled(ctx context.Context, staffID bson.ObjectID, monthStr string) error {
	filter := bson.M{
		"staffId":     staffID,
		"date":        bson.M{"$regex": "^" + monthStr},
		"status":      model.WorkShiftStatusScheduled,
		"actualHours": nil,
	}
	_, err := r.coll.DeleteMany(ctx, filter)
	return err
}

func (r *workShiftRepo) Update(ctx context.Context, id bson.ObjectID, updates bson.M) (*model.WorkShift, error) {
	updates["updatedAt"] = time.Now()
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var updated model.WorkShift
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *workShiftRepo) GetPayrollSummary(ctx context.Context, monthStr string) ([]model.MonthlyPayroll, error) {
	cursor, err := r.staffColl.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var staffList []model.StaffAccount
	if err := cursor.All(ctx, &staffList); err != nil {
		return nil, err
	}

	shifts, err := r.GetByMonth(ctx, monthStr, nil)
	if err != nil {
		return nil, err
	}

	payrollsCursor, err := r.payrollColl.Find(ctx, bson.M{"month": monthStr})
	if err != nil {
		return nil, err
	}
	defer payrollsCursor.Close(ctx)

	var savedPayrolls []model.MonthlyPayroll
	_ = payrollsCursor.All(ctx, &savedPayrolls)
	savedMap := make(map[string]model.MonthlyPayroll)
	for _, p := range savedMap {
		savedMap[p.StaffID.Hex()] = p
	}

	var results []model.MonthlyPayroll
	for _, s := range staffList {
		var totalSchedShifts, totalCompShifts, totalLeaveShifts int
		var totalSchedHours, totalActHours float64

		for _, shift := range shifts {
			if shift.StaffID == s.ID {
				totalSchedShifts++
				totalSchedHours += shift.ScheduledHours
				if shift.Status == model.WorkShiftStatusCompleted {
					totalCompShifts++
					if shift.ActualHours != nil {
						totalActHours += *shift.ActualHours
					} else {
						totalActHours += shift.ScheduledHours
					}
				} else if shift.Status == model.WorkShiftStatusLeaveExcused || shift.Status == model.WorkShiftStatusLeaveUnexcused {
					totalLeaveShifts++
				}
			}
		}

		hourlyRate := s.HourlyRate
		if hourlyRate == 0 {
			if s.Role == model.RoleOwner {
				hourlyRate = 50000
			} else {
				hourlyRate = 28000
			}
		}

		baseSalary := int64(totalActHours * float64(hourlyRate))
		saved, hasSaved := savedMap[s.ID.Hex()]
		var allowance, deduction int64
		var isFinalized bool
		var finAt *time.Time
		var finBy string

		if hasSaved {
			allowance = saved.Allowance
			deduction = saved.Deduction
			isFinalized = saved.IsFinalized
			finAt = saved.FinalizedAt
			finBy = saved.FinalizedBy
		}

		finalSalary := baseSalary + allowance - deduction
		if finalSalary < 0 {
			finalSalary = 0
		}

		results = append(results, model.MonthlyPayroll{
			Month:                monthStr,
			StaffID:              s.ID,
			StaffName:            s.Name,
			TotalScheduledShifts: totalSchedShifts,
			TotalCompletedShifts: totalCompShifts,
			TotalLeaveShifts:     totalLeaveShifts,
			TotalScheduledHours:  totalSchedHours,
			TotalActualHours:     totalActHours,
			HourlyRate:           hourlyRate,
			BaseSalary:           baseSalary,
			Allowance:            allowance,
			Deduction:            deduction,
			FinalSalary:          finalSalary,
			IsFinalized:          isFinalized,
			FinalizedAt:          finAt,
			FinalizedBy:          finBy,
		})
	}

	return results, nil
}

func (r *workShiftRepo) SavePayrollAdjustment(ctx context.Context, monthStr string, staffID bson.ObjectID, allowance, deduction int64) error {
	filter := bson.M{"month": monthStr, "staffId": staffID}
	update := bson.M{
		"$set": bson.M{
			"allowance": allowance,
			"deduction": deduction,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.payrollColl.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *workShiftRepo) FinalizePayroll(ctx context.Context, monthStr string, finalizedBy string) error {
	now := time.Now()
	filter := bson.M{"month": monthStr}
	update := bson.M{
		"$set": bson.M{
			"isFinalized": true,
			"finalizedAt": now,
			"finalizedBy": finalizedBy,
		},
	}
	opts := options.UpdateMany().SetUpsert(true)
	_, err := r.payrollColl.UpdateMany(ctx, filter, update, opts)
	return err
}
