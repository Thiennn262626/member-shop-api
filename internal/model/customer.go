package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CustomerTier string

const (
	TierMember  CustomerTier = "MEMBER"
	TierSilver  CustomerTier = "SILVER"
	TierGold    CustomerTier = "GOLD"
	TierDiamond CustomerTier = "DIAMOND"
)

type Customer struct {
	ID              bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Phone           string        `json:"phone" bson:"phone"` // Unique index
	Name            string        `json:"name" bson:"name"`
	Tier            CustomerTier  `json:"tier" bson:"tier"`
	EarnedPoints    int64         `json:"earnedPoints" bson:"earnedPoints"`
	RedeemedPoints  int64         `json:"redeemedPoints" bson:"redeemedPoints"`
	AvailablePoints int64         `json:"availablePoints" bson:"availablePoints"`
	TotalSpent      int64         `json:"totalSpent" bson:"totalSpent"`
	JoinedDate      string        `json:"joinedDate" bson:"joinedDate"`
	Version         int64         `json:"version" bson:"version"`
	CreatedAt       time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt" bson:"updatedAt"`
}

// CalculateTier returns the loyalty tier based on cumulative spent amount
func CalculateTier(totalSpent int64) CustomerTier {
	switch {
	case totalSpent >= 3000000:
		return TierDiamond
	case totalSpent >= 1200000:
		return TierGold
	case totalSpent >= 300000:
		return TierSilver
	default:
		return TierMember
	}
}

func GetTierMultiplier(tier CustomerTier) float64 {
	switch tier {
	case TierDiamond:
		return 1.5
	case TierGold:
		return 1.3
	case TierSilver:
		return 1.15
	default:
		return 1.0
	}
}
