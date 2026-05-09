package xauat

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"luminous/internal/util"
)

var examCache = util.NewCacheWithName("exam")
var examDataRegex = regexp.MustCompile(`var studentExamInfoVms = (.*?)\];`)
var hexEscapeRegex = regexp.MustCompile(`\\x[0-9A-Fa-f]{2}`)

// GetExams 获取考试安排
func GetExams(cookie, studentID string) (*ExamResponse, error) {
	cacheKey := fmt.Sprintf("exams:%s", studentID)
	val, err := examCache.GetOrSet(cacheKey, 1*time.Hour, func() (interface{}, error) {
		studentIDs := strings.Split(studentID, ",")
		var allExams []ExamInfo
		var mu sync.Mutex
		var wg sync.WaitGroup
		errCh := make(chan error, len(studentIDs))

		sem := make(chan struct{}, 8)
		for _, sid := range studentIDs {
			sid = strings.TrimSpace(sid)
			if sid == "" {
				continue
			}
			wg.Add(1)
			go func(id string) {
				sem <- struct{}{}
				defer func() { <-sem }()
				defer wg.Done()
				exams, err := fetchExamArrangement(cookie, id)
				if err != nil {
					errCh <- err
					return
				}
				mu.Lock()
				allExams = append(allExams, exams...)
				mu.Unlock()
			}(sid)
		}
		wg.Wait()
		close(errCh)

		var errs []error
		for err := range errCh {
			errs = append(errs, err)
		}

		if len(allExams) == 0 && len(errs) > 0 {
			return &ExamResponse{Exams: []ExamInfo{}, CanClick: false, Error: errs[0].Error()}, nil
		}
		for _, err := range errs {
			slog.Warn("partial exam fetch failed", "error", err)
		}

		return &ExamResponse{
			Exams:    allExams,
			CanClick: len(allExams) > 0,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*ExamResponse), nil
}

func fetchExamArrangement(cookie, studentID string) ([]ExamInfo, error) {
	url := baseURL + "/student/for-std/exam-arrange"
	if studentID != "" {
		url += "/info/" + studentID
	}

	body, err := fetchWithAuth(url, cookie)
	if err != nil {
		return nil, err
	}

	content := string(body)
	match := examDataRegex.FindStringSubmatch(content)
	if len(match) < 2 {
		return []ExamInfo{}, fmt.Errorf("failed to match exam data pattern")
	}

	jsonData := match[1] + "]"
	jsonData = strings.ReplaceAll(jsonData, "'", "\"")
	jsonData = hexEscapeRegex.ReplaceAllString(jsonData, "")

	var examData []examDataRaw
	if err := json.Unmarshal([]byte(jsonData), &examData); err != nil {
		return []ExamInfo{}, fmt.Errorf("parse exam json: %w", err)
	}

	exams := make([]ExamInfo, 0, len(examData))
	for _, d := range examData {
		exams = append(exams, ExamInfo{
			Name:     d.Course.NameZh,
			Time:     d.ExamTime,
			Location: d.Room,
			Seat:     d.SeatNo,
		})
	}
	return exams, nil
}
