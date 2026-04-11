package utils

// Struct utama untuk semua respons API
type APIResponse struct {
	Success bool        `json:"success"`
	Meta    Meta        `json:"meta"`
	Data    interface{} `json:"data,omitempty"`
}

type Meta struct {
	Message    string      `json:"message"`
	Code       int         `json:"code"`
	Status     string      `json:"status"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	LastPage    int   `json:"last_page"`
}

// BuildResponse untuk memformat respons sukses
func BuildResponse(message string, code int, status string, data interface{}) APIResponse {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  status,
	}

	return APIResponse{
		Success: true,
		Meta:    meta,
		Data:    data,
	}
}

// BuildResponseWithPagination untuk memformat respons sukses dengan pagination
func BuildResponseWithPagination(message string, code int, status string, data interface{}, pagination Pagination) APIResponse {
	meta := Meta{
		Message:    message,
		Code:       code,
		Status:     status,
		Pagination: &pagination,
	}

	return APIResponse{
		Success: true,
		Meta:    meta,
		Data:    data,
	}
}

// BuildErrorResponse untuk memformat exception/error dinamis
func BuildErrorResponse(message string, code int, status string, errors interface{}) APIResponse {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  status,
	}

	return APIResponse{
		Success: false,
		Meta:    meta,
		Data:    errors,
	}
}
