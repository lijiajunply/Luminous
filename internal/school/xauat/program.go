package xauat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"luminous/internal/util"
)

var programCache = util.NewCacheWithName("program")

// GetProgram 获取培养方案
func GetProgram(programID, cookie, filterName string) ([]PlanCourse, error) {
	cacheKey := fmt.Sprintf("program:%s", programID)
	val, err := programCache.GetOrSet(cacheKey, 24*time.Hour, func() (interface{}, error) {
		url := fmt.Sprintf(baseURL+"/student/for-std/program/root-module-json/%s", programID)
		body, err := fetchWithAuth(url, cookie)
		if err != nil {
			return nil, err
		}

		var root programModule
		if err := json.Unmarshal(body, &root); err != nil {
			return nil, fmt.Errorf("parse program json: %w", err)
		}

		return flattenModules(&root), nil
	})
	if err != nil {
		return nil, err
	}

	courses := val.([]PlanCourse)
	if filterName != "" {
		courses = filterByName(courses, filterName)
	}
	return courses, nil
}

func flattenModules(m *programModule) []PlanCourse {
	var result []PlanCourse
	result = append(result, m.PlanCourses...)
	for i := range m.Children {
		result = append(result, flattenModules(&m.Children[i])...)
	}
	return result
}

func filterByName(courses []PlanCourse, name string) []PlanCourse {
	var result []PlanCourse
	for _, c := range courses {
		if strings.Contains(c.Name, name) {
			result = append(result, c)
		}
	}
	return result
}

// GetProgramDict 按学期分组获取培养方案
func GetProgramDict(programID, cookie string) (map[string][]PlanCourse, error) {
	courses, err := GetProgram(programID, cookie, "")
	if err != nil {
		return nil, err
	}

	dict := make(map[string][]PlanCourse)
	for _, c := range courses {
		terms := strings.Split(c.TermStr, ",")
		for _, t := range terms {
			t = strings.TrimSpace(t)
			if t != "" {
				dict[t] = append(dict[t], c)
			}
		}
	}
	return dict, nil
}
