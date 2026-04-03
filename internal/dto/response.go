package model

type Meta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type BaseResponse struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
}

type AttendanceListResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}
