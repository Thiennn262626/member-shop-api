package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ItemOptionItem struct {
	ID          string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Price       int64  `json:"price" bson:"price"`
	IsDefault   bool   `json:"isDefault" bson:"isDefault"`
	IsAvailable bool   `json:"isAvailable" bson:"isAvailable"`
}

type ItemOptionGroup struct {
	ID        string           `json:"id" bson:"id"`
	Name      string           `json:"name" bson:"name"`
	Required  bool             `json:"required" bson:"required"`
	Selection string           `json:"selection" bson:"selection"` // 'only' | 'many'
	MinSelect int              `json:"minSelect,omitempty" bson:"minSelect,omitempty"`
	MaxSelect int              `json:"maxSelect,omitempty" bson:"maxSelect,omitempty"`
	Options   []ItemOptionItem `json:"options" bson:"options"`
}

type Product struct {
	ID                     bson.ObjectID     `json:"id" bson:"_id,omitempty"`
	Name                   string            `json:"name" bson:"name"`
	Category               string            `json:"category" bson:"category"`
	Description            string            `json:"description" bson:"description"`
	Price                  int64             `json:"price" bson:"price"`
	FakePrice              int64             `json:"fakePrice,omitempty" bson:"fakePrice,omitempty"`
	AllowPointRedemption   bool              `json:"allowPointRedemption" bson:"allowPointRedemption"`
	PointsRequiredToRedeem int64             `json:"pointsRequiredToRedeem,omitempty" bson:"pointsRequiredToRedeem,omitempty"`
	ImageURL               string            `json:"imageUrl" bson:"imageUrl"`
	IsActive               bool              `json:"isActive" bson:"isActive"`
	Unit                   string            `json:"unit" bson:"unit"`
	Tag                    string            `json:"tag,omitempty" bson:"tag,omitempty"`
	Rating                 float64           `json:"rating" bson:"rating"`
	SoldCount              int64             `json:"soldCount" bson:"soldCount"`
	OptionGroups           []ItemOptionGroup `json:"optionGroups,omitempty" bson:"optionGroups,omitempty"`
	CreatedAt              time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt              time.Time         `json:"updatedAt" bson:"updatedAt"`
}
