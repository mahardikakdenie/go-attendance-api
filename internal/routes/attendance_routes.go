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
		attendance.POST("/end-session", attendanceH.EndSession)

		corrections := attendance.Group("/corrections")
		{
			corrections.POST("", middleware.HasPermission("attendance.correction.create"), correctionH.RequestCorrection)
			corrections.GET("", middleware.HasPermission("attendance.correction.view"), correctionH.GetCorrections)
			corrections.POST("/:id/approve", middleware.HasPermission("attendance.correction.review"), correctionH.ApproveCorrection)
			corrections.POST("/:id/reject", middleware.HasPermission("attendance.correction.review"), correctionH.RejectCorrection)
		}
	}
}
