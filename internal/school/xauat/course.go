package xauat

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"luminous/internal/util"
)

var courseCache = util.NewCacheWithName("course")

// GetCourses 获取课表
func GetCourses(studentID, cookie string) (*CourseResultResponse, error) {
	cacheKey := "courses:" + studentID
	val, err := courseCache.GetOrSet(cacheKey, 24*time.Hour, func() (interface{}, error) {
		semester, err := GetCurrentSemester(cookie)
		if err != nil {
			return nil, fmt.Errorf("get current semester: %w", err)
		}
		if semester.Value == "" {
			return nil, fmt.Errorf("unable to determine current semester")
		}

		studentIDs := strings.Split(studentID, ",")
		var allCourses []CourseActivity
		var mu sync.Mutex
		var wg sync.WaitGroup
		errCh := make(chan error, len(studentIDs))

		sem := make(chan struct{}, 8)
		for i, sid := range studentIDs {
			sid = strings.TrimSpace(sid)
			if sid == "" {
				continue
			}
			wg.Add(1)
			go func(index int, id string) {
				sem <- struct{}{}
				defer func() { <-sem }()
				defer wg.Done()
				courses, err := fetchCoursesForStudent(semester.Value, id, cookie, index != 0)
				if err != nil {
					errCh <- fmt.Errorf("student %s: %w", id, err)
					return
				}
				mu.Lock()
				allCourses = append(allCourses, courses...)
				mu.Unlock()
			}(i, sid)
		}
		wg.Wait()
		close(errCh)

		var errs []error
		for err := range errCh {
			errs = append(errs, err)
		}

		if len(allCourses) == 0 && len(errs) > 0 {
			return nil, errs[0]
		}
		for _, err := range errs {
			slog.Warn("partial course fetch failed", "error", err)
		}

		for i := range allCourses {
			sort.Ints(allCourses[i].WeekIndexes)
			if allCourses[i].Room == "" {
				allCourses[i].Room = "未知"
			}
			allCourses[i].Room = strings.ReplaceAll(allCourses[i].Room, "*", "")
		}

		return &CourseResultResponse{
			Success:        true,
			Data:           allCourses,
			ExpirationTime: time.Now().Add(24 * time.Hour),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*CourseResultResponse), nil
}

func fetchCoursesForStudent(semesterValue, studentID, cookie string, isMinor bool) ([]CourseActivity, error) {
	url := fmt.Sprintf(baseURL+"/student/for-std/course-table/semester/%s/print-data/%s", semesterValue, studentID)
	body, err := fetchWithAuth(url, cookie)
	if err != nil {
		return nil, err
	}

	var cr courseResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("parse course json: %w", err)
	}

	return cr.StudentTableVm.Activities, nil
}
