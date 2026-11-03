package apperror

import (
	"fmt"
	"net/http"
)

// AppError đại diện cho lỗi nghiệp vụ chuẩn hóa trong toàn hệ thống.
type AppError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	Err        error  `json:"-"` // Lỗi gốc bên dưới (dùng để log, không leak ra client)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

func (e *AppError) WithErr(err error) *AppError {
	e.Err = err
	return e
}

// Bảng mã lỗi tiêu chuẩn (Error Codes)
const (
	// System / Protocol Codes
	CodeSuccess             = "SUCCESS"
	CodeBadRequest          = "BAD_REQUEST"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	CodeConflict            = "CONFLICT"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"

	// Auth & Permission Codes
	CodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeAuthAccountLocked      = "AUTH_ACCOUNT_LOCKED"
	CodeAuthTokenExpired       = "AUTH_TOKEN_EXPIRED"
	CodeAuthTokenInvalid       = "AUTH_TOKEN_INVALID"

	// Order & Store Domain Codes
	CodeOrderNotFound      = "ORDER_NOT_FOUND"
	CodeOrderInvalidStatus = "ORDER_INVALID_STATUS"
	CodeProductNotFound    = "PRODUCT_NOT_FOUND"
	CodeCustomerNotFound   = "CUSTOMER_NOT_FOUND"
	CodeInsufficientPoints = "INSUFFICIENT_POINTS"

	// Voucher Domain Codes
	CodeVoucherNotFound       = "VOUCHER_NOT_FOUND"
	CodeVoucherExpired        = "VOUCHER_EXPIRED"
	CodeVoucherInactive       = "VOUCHER_INACTIVE"
	CodeVoucherLimitReached   = "VOUCHER_LIMIT_REACHED"
	CodeVoucherMinSpendNotMet = "VOUCHER_MIN_SPEND_NOT_MET"

	// Shift Domain Codes
	CodeShiftAlreadyOpen = "SHIFT_ALREADY_OPEN"
	CodeShiftNotFound    = "SHIFT_NOT_FOUND"
	CodeShiftNotOpen     = "SHIFT_NOT_OPEN"
	CodeShiftAlreadyClosed = "SHIFT_ALREADY_CLOSED"

	// Work Shift & Staff Domain Codes
	CodeStaffNotFound             = "STAFF_NOT_FOUND"
	CodeWorkShiftNotFound         = "WORK_SHIFT_NOT_FOUND"
	CodeWorkShiftAlreadyConfirmed = "WORK_SHIFT_ALREADY_CONFIRMED"
)

// Khởi tạo AppError tùy chỉnh
func New(httpStatus int, code string, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

// Bọc lỗi gốc (underlying error) kèm HTTPStatus, Code và Message
func Wrap(err error, httpStatus int, code string, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// Các hàm tiện ích phổ biến
func BadRequest(message string, details ...any) *AppError {
	err := New(http.StatusBadRequest, CodeBadRequest, message)
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

func ValidationError(message string, details any) *AppError {
	err := New(http.StatusBadRequest, CodeValidationError, message)
	err.Details = details
	return err
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

func Internal(err error, message ...string) *AppError {
	msg := "Hệ thống gặp sự cố ngoài dự kiến, vui lòng thử lại sau"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return Wrap(err, http.StatusInternalServerError, CodeInternalServerError, msg)
}
