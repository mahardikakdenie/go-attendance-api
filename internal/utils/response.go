package utils

// Struct utama untuk semua respons API
type APIResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data,omitempty"`
}

type Meta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

// BuildResponse untuk memformat respons sukses
func BuildResponse(message string, code int, status string, data interface{}) APIResponse {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  status,
	}

	return APIResponse{
		Meta: meta,
		Data: data,
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
		Meta: meta,
		Data: errors,
	}
}
