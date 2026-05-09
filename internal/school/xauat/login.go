package xauat

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"luminous/internal/util"
)

var studentIDRegex = regexp.MustCompile(`value="(\d+)"`)

// Login 执行 SSO 登录，返回学生 ID 和 Cookie
func Login(username, password string) (*LoginResponse, error) {
	fullURL := fmt.Sprintf("%s/login/%s/%s", loginURL, username, password)
	resp, err := util.DefaultClient.GetWithCookie(fullURL, "")
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read login response: %w", err)
	}

	var result struct {
		Success bool   `json:"success"`
		Cookies string `json:"cookies"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse login response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("login failed: invalid credentials")
	}

	cookie := parseCookie(result.Cookies)
	studentID, err := getStudentID(cookie)
	if err != nil {
		return nil, fmt.Errorf("get student ID: %w", err)
	}

	return &LoginResponse{
		Success:   true,
		StudentID: studentID,
		Cookie:    cookie,
	}, nil
}

func getStudentID(cookie string) (string, error) {
	resp, err := util.DefaultClient.GetWithCookie(baseURL+"/student/for-std/precaution", cookie)
	if err != nil {
		return "", fmt.Errorf("precaution request: %w", err)
	}
	defer resp.Body.Close()

	// Check if redirected to a student-specific path
	path := resp.Request.URL.Path
	if path != "/student/for-std/precaution" {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		for i, p := range parts {
			if p == "index" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		return strings.TrimPrefix(path, "/student/for-std/precaution/index/"), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read precaution body: %w", err)
	}

	if strings.Contains(string(body), "登入页面") {
		return "", fmt.Errorf("authentication failed: cookie expired or invalid")
	}

	matches := studentIDRegex.FindStringSubmatch(string(body))
	if len(matches) >= 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("unable to extract student ID from precaution page")
}

// parseCookie 提取 __pstsid__ 和 SESSION cookie
func parseCookie(cookies string) string {
	var result []string
	for _, c := range strings.Split(cookies, ";") {
		c = strings.TrimSpace(c)
		if strings.Contains(c, "__pstsid__") {
			result = append(result, c)
		} else if strings.Contains(c, "SESSION") {
			parts := strings.SplitN(c, "=", 2)
			if len(parts) == 2 {
				result = append(result, "SESSION="+strings.SplitN(parts[1], ";", 2)[0])
			}
		}
	}
	return strings.Join(result, ";")
}
