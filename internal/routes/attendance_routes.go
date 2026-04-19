package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAttendanceRoutes(rg *gin.RouterGroup, attendanceH handler.AttendanceHandler, correctionH handler.AttendanceCorrectionHandler) {
	attendance := rg.Group("/attendance")
	{
		attendance.POST("", attendanceH.RecordAttendance)
		attendance.GET("", attendanceH.GetAllAttendance)
		attendance.GET("/history", attendanceH.GetAttendanceHistory)
		attendance.GET("/summary", attendanceH.GetAttendanceSummary)
		attendance.GET("/today", attendanceH.GetTodayAttendance)

		corrections := attendance.Group("/corrections")
		{
			corrections.POST("", correctionH.RequestCorrection)
			corrections.GET("", correctionH.GetCorrections)
			corrections.POST("/:id/approve", middleware.RequireRole("superadmin", "admin", "hr"), correctionH.ApproveCorrection)
			corrections.POST("/:id/reject", middleware.RequireRole("superadmin", "admin", "hr"), correctionH.RejectCorrection)
		}
	}
}
