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

var scoreCache = util.NewCacheWithName("score")
var spanRegex = regexp.MustCompile(`<span[^>]*>([^<]+)</span>`)

// GetScores 获取成绩
func GetScores(studentID, semester, cookie string) ([]ScoreItem, error) {
	cacheKey := fmt.Sprintf("scores:%s:%s", studentID, semester)
	val, err := scoreCache.GetOrSet(cacheKey, 1*time.Hour, func() (interface{}, error) {
		studentIDs := strings.Split(studentID, ",")
		var allScores []ScoreItem
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
				scores, err := crawlScores(id, semester, cookie, index != 0)
				if err != nil {
					errCh <- err
					return
				}
				mu.Lock()
				allScores = append(allScores, scores...)
				mu.Unlock()
			}(i, sid)
		}
		wg.Wait()
		close(errCh)

		var errs []error
		for err := range errCh {
			errs = append(errs, err)
		}

		if len(allScores) == 0 && len(errs) > 0 {
			return nil, errs[0]
		}
		for _, err := range errs {
			slog.Warn("partial score fetch failed", "error", err)
		}
		return allScores, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]ScoreItem), nil
}

func crawlScores(studentID, semester, cookie string, isMinor bool) ([]ScoreItem, error) {
	url := fmt.Sprintf(baseURL+"/student/for-std/grade/sheet/info/%s?semester=%s", studentID, semester)
	body, err := fetchWithAuth(url, cookie)
	if err != nil {
		return nil, err
	}

	content := string(body)
	if strings.HasPrefix(content, "<") {
		return nil, fmt.Errorf("unexpected HTML response from score endpoint for student %s", studentID)
	}

	var rawJSON map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		return nil, fmt.Errorf("parse score json: %w", err)
	}

	gradesData, ok := rawJSON["semesterId2studentGrades"]
	if !ok {
		return []ScoreItem{}, nil
	}

	var gradesMap map[string][]struct {
		Course struct {
			NameZh  string `json:"nameZh"`
			Credits string `json:"credits"`
		} `json:"course"`
		LessonCode   string `json:"lessonCode"`
		LessonNameZh string `json:"lessonNameZh"`
		GaGrade      string `json:"gaGrade"`
		Gp           string `json:"gp"`
		GradeDetail  string `json:"gradeDetail"`
	}
	if err := json.Unmarshal(gradesData, &gradesMap); err != nil {
		return nil, fmt.Errorf("parse grades map: %w", err)
	}

	semesterGrades, ok := gradesMap[semester]
	if !ok {
		return []ScoreItem{}, nil
	}

	scores := make([]ScoreItem, 0, len(semesterGrades))
	for _, g := range semesterGrades {
		detailParts := spanRegex.FindAllStringSubmatch(g.GradeDetail, -1)
		detailTexts := make([]string, 0, len(detailParts))
		for _, d := range detailParts {
			if len(d) >= 2 {
				detailTexts = append(detailTexts, strings.TrimSpace(d[1]))
			}
		}

		scores = append(scores, ScoreItem{
			Name:        g.Course.NameZh,
			LessonCode:  g.LessonCode,
			LessonName:  g.LessonNameZh,
			Grade:       g.GaGrade,
			GPA:         g.Gp,
			GradeDetail: strings.Join(detailTexts, "; "),
			Credit:      g.Course.Credits,
			IsMinor:     isMinor,
		})
	}

	return scores, nil
}
