package xauat

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"luminous/internal/util"
)

const completionPath = "/student/ws/student/home-page/programCompletionPreview"

var infoCache = util.NewCacheWithName("info")
var semesterDateCache = util.NewCacheWithName("semester_date")

// GetCompletion 获取学业完成进度
func GetCompletion(cookie string) ([]StudyModule, error) {
	cacheKey := "completion"
	val, err := infoCache.GetOrSet(cacheKey, 1*time.Hour, func() (interface{}, error) {
		body, err := fetchWithAuth(baseURL+completionPath, cookie)
		if err != nil {
			return nil, err
		}

		var modules []StudyModule
		if err := json.Unmarshal(body, &modules); err != nil {
			return nil, fmt.Errorf("parse completion json: %w", err)
		}
		return modules, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]StudyModule), nil
}

// GetTimeInfo 获取学期时间范围。若提供有效 cookie 则尝试从当前学期推算日期。
func GetTimeInfo(cookie string) map[string]string {
	val, err := semesterDateCache.GetOrSet("semester_dates", 24*time.Hour, func() (interface{}, error) {
		dates := map[string]string{
			"start_time": semesterStart,
			"end_time":   semesterEnd,
		}
		if cookie != "" {
			semester, err := GetCurrentSemester(cookie)
			if err != nil {
				slog.Warn("cannot fetch current semester, using config dates", "error", err)
			} else if start, end, ok := parseSemesterDates(semester.Text); ok {
				dates["start_time"] = start
				dates["end_time"] = end
			} else {
				slog.Warn("cannot parse semester text, using config dates", "text", semester.Text)
			}
		}
		return dates, nil
	})
	if err != nil {
		return map[string]string{
			"start_time": semesterStart,
			"end_time":   semesterEnd,
		}
	}
	return val.(map[string]string)
}

func parseSemesterDates(text string) (string, string, bool) {
	parts := strings.Split(text, "-")
	if len(parts) != 3 {
		return "", "", false
	}
	y1, e1 := strconv.Atoi(parts[0])
	y2, e2 := strconv.Atoi(parts[1])
	sem, e3 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || e3 != nil {
		return "", "", false
	}
	if sem == 1 {
		return fmt.Sprintf("%d-09-01", y1), fmt.Sprintf("%d-01-15", y2), true
	}
	return fmt.Sprintf("%d-02-25", y2), fmt.Sprintf("%d-07-15", y2), true
}
