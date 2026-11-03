package dto

type ValidateVoucherRequest struct {
	Code     string `json:"code" binding:"required"`
	SubTotal int64  `json:"subTotal" binding:"required,min=0"`
}
