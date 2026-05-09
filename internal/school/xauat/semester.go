package xauat

import (
	"fmt"
	"regexp"
	"strings"

	"luminous/internal/util"
)

var optionRegex = regexp.MustCompile(`<option(?:\s+selected="selected")?\s+value="([^"]*)">([^<]*)</option>`)

// fetchWithAuth 发起带 Cookie 的 GET 请求，自动检查认证
func fetchWithAuth(url, cookie string) ([]byte, error) {
	body, err := util.DefaultClient.FetchWithCookie(url, cookie)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	if strings.Contains(string(body), "登入页面") {
		return nil, &AuthError{Message: "authentication failed: cookie expired"}
	}
	return body, nil
}

// ParseSemesters 解析学期列表
func ParseSemesters(cookie, studentID string) (*SemesterResult, error) {
	url := baseURL + "/student/for-std/grade/sheet"
	if studentID != "" {
		url += "/semester-index/" + studentID
	}

	body, err := fetchWithAuth(url, cookie)
	if err != nil {
		return nil, err
	}

	result := &SemesterResult{Data: []SemesterItem{}}
	matches := optionRegex.FindAllStringSubmatch(string(body), -1)
	for _, m := range matches {
		result.Data = append(result.Data, SemesterItem{
			Value: m[1],
			Text:  m[2],
		})
	}

	return result, nil
}

// GetCurrentSemester 获取当前学期
func GetCurrentSemester(cookie string) (*SemesterItem, error) {
	body, err := fetchWithAuth(baseURL+"/student/for-std/course-table", cookie)
	if err != nil {
		return nil, err
	}

	html := string(body)
	selectedRegex := regexp.MustCompile(`<option\s+selected="selected"\s+value="([^"]*)">([^<]*)</option>`)
	match := selectedRegex.FindStringSubmatch(html)
	if len(match) >= 3 {
		return &SemesterItem{Value: match[1], Text: match[2]}, nil
	}

	matches := optionRegex.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		m := matches[0]
		return &SemesterItem{Value: m[1], Text: m[2]}, nil
	}

	return &SemesterItem{Value: "301", Text: "2025-2026-1"}, nil
}

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
