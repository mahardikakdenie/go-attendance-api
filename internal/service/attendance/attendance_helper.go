package attendance

import (
	"math"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"
)

var allowedAttendanceIncludes = map[string]bool{
	"user":    true,
	"tenant":  true,
	"setting": true,
}

// filterAttendanceIncludes memfilter array string include dari request agar hanya menyisakan relasi yang diizinkan (whitelist).
//
// Parameters:
//   - includes: []string, array string relasi dari request.
//
// Returns/Impact:
//   - []string: array string yang sudah bersih dari input berbahaya/tidak valid.
func filterAttendanceIncludes(includes []string) []string {
	var result []string
	for _, inc := range includes {
		if allowedAttendanceIncludes[inc] {
			result = append(result, inc)
		}
	}
	return result
}

// hasAttendanceInclude mengecek apakah suatu kunci (key) relasi ada di dalam daftar include.
//
// Parameters:
//   - includes: []string, array yang ingin dicek.
//   - key: string, relasi yang dicari.
//
// Returns/Impact:
//   - bool: true jika ditemukan, false jika tidak.
func hasAttendanceInclude(includes []string, key string) bool {
	for _, inc := range includes {
		if inc == key {
			return true
		}
	}
	return false
}

// calculateDiff menghitung persentase perbedaan antara nilai periode saat ini dengan periode sebelumnya.
//
// Parameters:
//   - today: int64, nilai total hari ini / periode sekarang.
//   - previous: int64, nilai total periode lalu.
//
// Returns/Impact:
//   - float64: persentase kenaikan (positif) atau penurunan (negatif). Jika previous 0, bernilai 100 atau 0.
func calculateDiff(today, previous int64) float64 {
	if previous == 0 {
		if today > 0 {
			return 100.0
		}
		return 0.0
	}
	return float64(today-previous) / float64(previous) * 100.0
}

// isWithinTimeRange memvalidasi apakah waktu saat ini berada di antara batas waktu mulai dan selesai.
// Mendukung validasi waktu yang melewati tengah malam (misal shift malam: 22:00 - 06:00).
//
// Parameters:
//   - now: time.Time, waktu saat ini (sudah dilock ke WIB).
//   - start, end: string, format "15:04" batas waktu mulai dan selesai.
//
// Returns/Impact:
//   - bool: true jika masuk dalam range waktu.
//   - error: jika gagal memparsing format jam start/end.
func isWithinTimeRange(now time.Time, start, end string) (bool, error) {
	layout := "15:04"

	if start == "24:00" {
		start = "00:00"
	}
	if end == "24:00" {
		end = "23:59"
	}

	startTime, err := utils.ParseTimeWIB(layout, start)
	if err != nil {
		return false, err
	}

	endTime, err := utils.ParseTimeWIB(layout, end)
	if err != nil {
		return false, err
	}

	// now.Hour() dan now.Minute() di sini akan mengekstrak jam lokal (UTC+7)
	// berkat `.In(WIB)` yang kita aplikasikan di atas.
	current := now.Hour()*60 + now.Minute()
	startMin := startTime.Hour()*60 + startTime.Minute()
	endMin := endTime.Hour()*60 + endTime.Minute()

	if startMin <= endMin {
		return current >= startMin && current <= endMin, nil
	}

	// Kasus jika shift malam (misal start 22:00, end 06:00)
	return current >= startMin || current <= endMin, nil
}

// calculateDistance menghitung jarak antara dua titik koordinat latitude dan longitude menggunakan rumus Haversine.
//
// Parameters:
//   - lat1, lon1: float64, koordinat titik referensi (kantor).
//   - lat2, lon2: float64, koordinat titik user saat ini.
//
// Returns/Impact:
//   - float64: jarak dalam satuan meter.
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*
			math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// normalizeTime menormalkan string format waktu. Mengubah "24:00" menjadi "23:59".
//
// Parameters:
//   - t: string, waktu format "15:04".
//
// Returns/Impact:
//   - string: waktu yang sudah dinormalkan.
func normalizeTime(t string) string {
	if t == "24:00" {
		return "23:59"
	}
	return t
}

// applyPreloads memetakan entitas database (model.Attendance) ke DTO respons (model.AttendanceResponse).
// Ini termasuk memetakan log absensi dan relasi user jika diminta.
//
// Parameters:
//   - a: *model.Attendance, entitas database sumber.
//   - includes: []string, list relasi yang ingin diekstrak (contoh: "user").
//
// Returns/Impact:
//   - model.AttendanceResponse: objek respons bersih siap di-serve ke client.
func applyPreloads(a *model.Attendance, includes []string) model.AttendanceResponse {
	res := model.AttendanceResponse{
		ID:                a.ID,
		UserID:            a.UserID,
		ClockInTime:       a.ClockInTime,
		ClockOutTime:      a.ClockOutTime,
		ClockInLatitude:   a.ClockInLatitude,
		ClockInLongitude:  a.ClockInLongitude,
		ClockOutLatitude:  a.ClockOutLatitude,
		ClockOutLongitude: a.ClockOutLongitude,
		ClockInMediaUrl:   a.ClockInMediaUrl,
		ClockOutMediaUrl:  a.ClockOutMediaUrl,
		Status:            a.Status,
		CreatedAt:         a.ClockInTime, // Using ClockInTime as CreatedAt
	}

	// Map AttendanceLog entries to AttendanceLogResponse.
	for _, l := range a.Logs {
		res.Logs = append(res.Logs, model.AttendanceLogResponse{
			ID:           l.ID,
			AttendanceID: l.AttendanceID,
			Action:       l.Action,
			LogTime:      l.LogTime,
			Latitude:     l.Latitude,
			Longitude:    l.Longitude,
			MediaUrl:     l.MediaUrl,
		})
	}

	if hasAttendanceInclude(includes, "user") {
		var roleRes *model.RoleResponse
		if a.User.Role != nil {
			roleRes = &model.RoleResponse{
				ID:   a.User.Role.ID,
				Name: a.User.Role.Name,
			}
		}

		res.User = &model.UserResponse{
			ID:          a.User.ID,
			Name:        a.User.Name,
			Email:       a.User.Email,
			Role:        roleRes,
			TenantID:    a.User.TenantID,
			EmployeeID:  a.User.EmployeeID,
			Department:  a.User.Department,
			MediaUrl:    a.User.MediaUrl,
			Address:     a.User.Address,
			PhoneNumber: a.User.PhoneNumber,
			CreatedAt:   a.User.CreatedAt,
		}
	}

	return res
}

// resolveAllowMultipleCheck menentukan apakah fitur multiple check (absen lebih dari sekali sehari) diizinkan.
// Prioritas cek adalah userSetting terlebih dahulu, jika nil maka akan fallback ke tenantSetting.
//
// Parameters:
//   - tenantSetting: *model.TenantSetting, pengaturan global tingkat tenant.
//   - userSetting: *model.UserSetting, pengaturan spesifik per user.
//
// Returns/Impact:
//   - bool: true jika diperbolehkan, false jika dilarang.
func resolveAllowMultipleCheck(tenantSetting *model.TenantSetting, userSetting *model.UserSetting) bool {
	if userSetting != nil && userSetting.AllowMultipleCheck != nil {
		return *userSetting.AllowMultipleCheck
	}
	return tenantSetting.AllowMultipleCheck
}
