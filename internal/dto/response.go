package modelDto

type Meta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type PaginationMeta struct {
	Total       int64 `json:"total"`
	CurrentPage int   `json:"current_page"`
	LastPage    int   `json:"last_page"`
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

type AttendanceHistoryItem struct {
	ID       string `json:"id"`
	Employee struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"employee"`
	Date     string `json:"date"`
	ClockIn  string `json:"clock_in"`
	ClockOut string `json:"clock_out"`
	Location string `json:"location"`
	Status   string `json:"status"`
	Overtime string `json:"overtime"`
}

type AttendanceHistoryResponse struct {
	Success bool                    `json:"success"`
	Data    []AttendanceHistoryItem `json:"data"`
	Meta    PaginationMeta          `json:"meta"`
}

type QuickInfoResponse struct {
	PendingLeaves      int `json:"pending_leaves"`
	PendingOvertimes   int `json:"pending_overtimes"`
	NotificationsCount int `json:"notifications_count"`
}

type TodayAttendanceResponse struct {
	ClockInTime  string `json:"clock_in_time"`
	ClockOutTime string `json:"clock_out_time"`
	Status       string `json:"status"`
	Duration     string `json:"duration"`
	Date         string `json:"date"`
}
