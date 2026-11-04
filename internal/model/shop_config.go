package model

type BankConfig struct {
	BankName      string `json:"bankName" bson:"bankName"`
	AccountNumber string `json:"accountNumber" bson:"accountNumber"`
	AccountHolder string `json:"accountHolder" bson:"accountHolder"`
}

type ShippingTier struct {
	ID       string `json:"id" bson:"id"`
	MinOrder int64  `json:"minOrder" bson:"minOrder"`
	MaxOrder *int64 `json:"maxOrder,omitempty" bson:"maxOrder,omitempty"`
	Fee      int64  `json:"fee" bson:"fee"`
	Label    string `json:"label" bson:"label"`
}

type DeliveryConfig struct {
	Enabled       bool           `json:"enabled" bson:"enabled"`
	MaxDistanceKm int            `json:"maxDistanceKm" bson:"maxDistanceKm"`
	AreaNote      string         `json:"areaNote" bson:"areaNote"`
	Tiers         []ShippingTier `json:"tiers" bson:"tiers"`
}

type ShopConfig struct {
	ID                string         `json:"-" bson:"_id"` // Ẩn khỏi JSON để không trả dư trường ID cấu hình nội bộ
	ShopName          string         `json:"shopName" bson:"shopName"`
	Tagline           string         `json:"tagline" bson:"tagline"`
	Hotline           string         `json:"hotline" bson:"hotline"`
	ZaloPhone         string         `json:"zaloPhone" bson:"zaloPhone"`
	Address           string         `json:"address" bson:"address"`
	PointToVndRatio   int64          `json:"pointToVndRatio" bson:"pointToVndRatio"`
	SpendToPointRatio int64          `json:"spendToPointRatio" bson:"spendToPointRatio"`
	IsOpen            bool           `json:"isOpen" bson:"isOpen"`
	PauseReason       string         `json:"pauseReason,omitempty" bson:"pauseReason,omitempty"`
	Bank              BankConfig     `json:"bank" bson:"bank"`
	Delivery          DeliveryConfig `json:"delivery" bson:"delivery"`
	Version           int64          `json:"version" bson:"version"`
}

func (s *ShopConfig) CalculateShippingFee(subTotal int64, fulfillmentType string) int64 {
	if fulfillmentType == "pickup" || !s.Delivery.Enabled {
		return 0
	}
	for _, tier := range s.Delivery.Tiers {
		if subTotal >= tier.MinOrder && (tier.MaxOrder == nil || subTotal <= *tier.MaxOrder) {
			return tier.Fee
		}
	}
	return 20000 // default fallback
}
