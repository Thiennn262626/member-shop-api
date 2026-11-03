package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/logger"
	"net/http"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Response đại diện cho định dạng phản hồi JSON chuẩn hóa của toàn bộ hệ thống API.
type Response struct {
	Success   bool      `json:"success"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Data      any       `json:"data,omitempty"`
	Meta      any       `json:"meta,omitempty"`
	Errors    any       `json:"errors,omitempty"`
	TraceID   string    `json:"traceId,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Meta đại diện cho thông tin phân trang chuẩn.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"pageSize,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"totalPages,omitempty"`
}

// FieldError đại diện cho chi tiết lỗi xác thực của từng trường dữ liệu.
type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message"`
}

// GetTraceID trích xuất traceId từ gin.Context hoặc header request.
func GetTraceID(c *gin.Context) string {
	if traceID := c.GetString(string(logger.TraceIdKey)); traceID != "" {
		return traceID
	}
	if traceID := c.GetHeader("X-Trace-Id"); traceID != "" {
		return traceID
	}
	if c.Request != nil {
		return logger.GetTraceID(c.Request.Context())
	}
	return ""
}

// Success trả về phản hồi thành công chuẩn HTTP 200 OK.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Code:      apperror.CodeSuccess,
		Message:   "Thao tác thành công",
		Data:      data,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// SuccessWithMsg trả về phản hồi thành công HTTP 200 OK kèm thông điệp tùy chỉnh.
func SuccessWithMsg(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Code:      apperror.CodeSuccess,
		Message:   message,
		Data:      data,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// Created trả về phản hồi thành công tạo tài nguyên mới HTTP 201 Created.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Code:      apperror.CodeSuccess,
		Message:   "Tạo mới thành công",
		Data:      data,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// CreatedWithMsg trả về phản hồi HTTP 201 Created kèm thông điệp tùy chỉnh.
func CreatedWithMsg(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Code:      apperror.CodeSuccess,
		Message:   message,
		Data:      data,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// SuccessWithMeta trả về phản hồi thành công có kèm dữ liệu phân trang/metadata.
func SuccessWithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Code:      apperror.CodeSuccess,
		Message:   "Thao tác thành công",
		Data:      data,
		Meta:      meta,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// Fail trả về phản hồi lỗi theo đối tượng AppError chuẩn hóa.
func Fail(c *gin.Context, appErr *apperror.AppError) {
	status := appErr.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}

	c.JSON(status, Response{
		Success:   false,
		Code:      appErr.Code,
		Message:   appErr.Message,
		Errors:    appErr.Details,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// FailWithStatus trả về phản hồi lỗi nhanh theo mã HTTP status, Code và Message.
func FailWithStatus(c *gin.Context, httpStatus int, code string, message string) {
	c.JSON(httpStatus, Response{
		Success:   false,
		Code:      code,
		Message:   message,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// FailWithDetailed trả về phản hồi lỗi kèm chi tiết (tương thích mẫu template NoRouteHandler / NoMethodHandler).
func FailWithDetailed(c *gin.Context, httpStatus int, errors any, message string) {
	code := httpStatusToCode(httpStatus)
	c.JSON(httpStatus, Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Errors:    errors,
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// HandleValidationError xử lý lỗi binding và validation khi nhận request body/query.
func HandleValidationError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		fieldErrors := make([]FieldError, 0, len(valErr))
		for _, f := range valErr {
			fieldName := toCamelCase(f.Field())
			fieldErrors = append(fieldErrors, FieldError{
				Field:   fieldName,
				Rule:    f.Tag(),
				Value:   f.Value(),
				Message: formatValidationMessage(fieldName, f.Tag(), f.Param()),
			})
		}
		c.JSON(http.StatusBadRequest, Response{
			Success:   false,
			Code:      apperror.CodeValidationError,
			Message:   "Dữ liệu gửi lên không đáp ứng quy chuẩn hợp lệ",
			Errors:    fieldErrors,
			TraceID:   GetTraceID(c),
			Timestamp: time.Now().UTC(),
		})
		return
	}

	var jsonErr *json.UnmarshalTypeError
	if errors.As(err, &jsonErr) {
		msg := fmt.Sprintf("Trường '%s' yêu cầu kiểu dữ liệu '%s' nhưng nhận được '%s'",
			jsonErr.Field, jsonErr.Type.String(), jsonErr.Value)
		c.JSON(http.StatusBadRequest, Response{
			Success:   false,
			Code:      apperror.CodeBadRequest,
			Message:   msg,
			TraceID:   GetTraceID(c),
			Timestamp: time.Now().UTC(),
		})
		return
	}

	if errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, Response{
			Success:   false,
			Code:      apperror.CodeBadRequest,
			Message:   "Nội dung yêu cầu (request body) không được để trống",
			TraceID:   GetTraceID(c),
			Timestamp: time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusBadRequest, Response{
		Success:   false,
		Code:      apperror.CodeBadRequest,
		Message:   err.Error(),
		TraceID:   GetTraceID(c),
		Timestamp: time.Now().UTC(),
	})
}

// HandleError là bộ phân luồng xử lý lỗi thông minh cho tất cả handler:
// Tự nhận diện AppError, ValidationErrors, Mongo NoDocument hoặc lỗi không mong muốn (500).
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 1. Lỗi nghiệp vụ AppError đã được định nghĩa
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		Fail(c, appErr)
		return
	}

	// 2. Lỗi validation của Gin / Playground Validator
	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		HandleValidationError(c, valErr)
		return
	}

	// 3. Lỗi không tìm thấy bản ghi trong MongoDB
	if errors.Is(err, mongo.ErrNoDocuments) {
		FailWithStatus(c, http.StatusNotFound, apperror.CodeNotFound, "Không tìm thấy dữ liệu yêu cầu")
		return
	}

	// 4. Lỗi hệ thống bất ngờ: Ghi log chi tiết (kèm traceId, path, method, err) nhưng ẩn chi tiết nội bộ với client
	traceID := GetTraceID(c)
	if logger.Log != nil {
		logger.Log.Error().
			Str("trace_id", traceID).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Err(err).
			Msg("Unhandled internal server error")
	}

	FailWithStatus(c, http.StatusInternalServerError, apperror.CodeInternalServerError, "Hệ thống gặp sự cố ngoài dự kiến, vui lòng thử lại sau.")
}

// Chuyển đổi mã HTTP Status thành Code String chuẩn
func httpStatusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return apperror.CodeBadRequest
	case http.StatusUnauthorized:
		return apperror.CodeUnauthorized
	case http.StatusForbidden:
		return apperror.CodeForbidden
	case http.StatusNotFound:
		return apperror.CodeNotFound
	case http.StatusMethodNotAllowed:
		return apperror.CodeMethodNotAllowed
	case http.StatusConflict:
		return apperror.CodeConflict
	case http.StatusTooManyRequests:
		return apperror.CodeTooManyRequests
	default:
		return apperror.CodeInternalServerError
	}
}

// Định dạng thông báo lỗi xác thực tiếng Việt dễ hiểu
func formatValidationMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("Trường '%s' là bắt buộc", field)
	case "email":
		return fmt.Sprintf("Trường '%s' phải là địa chỉ email hợp lệ", field)
	case "min":
		return fmt.Sprintf("Trường '%s' phải có tối thiểu %s ký tự hoặc phần tử", field, param)
	case "max":
		return fmt.Sprintf("Trường '%s' không được vượt quá %s ký tự hoặc phần tử", field, param)
	case "gt":
		return fmt.Sprintf("Trường '%s' phải lớn hơn %s", field, param)
	case "gte":
		return fmt.Sprintf("Trường '%s' phải lớn hơn hoặc bằng %s", field, param)
	case "lt":
		return fmt.Sprintf("Trường '%s' phải nhỏ hơn %s", field, param)
	case "lte":
		return fmt.Sprintf("Trường '%s' phải nhỏ hơn hoặc bằng %s", field, param)
	case "len":
		return fmt.Sprintf("Trường '%s' phải có độ dài chính xác là %s", field, param)
	case "oneof":
		return fmt.Sprintf("Trường '%s' phải là một trong các giá trị: [%s]", field, param)
	default:
		return fmt.Sprintf("Trường '%s' không thỏa mãn điều kiện quy chuẩn (%s)", field, tag)
	}
}

// Chuyển chữ hoa đầu dòng sang lowerCamelCase (ví dụ: CustomerPhone -> customerPhone)
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
