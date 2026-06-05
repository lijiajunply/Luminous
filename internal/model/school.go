package model

import (
	"net"
	"net/url"
	"regexp"
	"time"
)

type Feature string

const (
	FeatureTimetable     Feature = "timetable"       // 日历功能
	FeatureGradeQuery    Feature = "grade_query"     // 成绩查询
	FeatureGPACalc       Feature = "gpa_calculation" // GPA计算，需要成绩查询功能支持
	FeatureCourseSelect  Feature = "course_schedule" // 课程显示
	FeatureExamSchedule  Feature = "exam_schedule"   // 考试安排
	FeatureLogin         Feature = "login"           // 登录，最基础服务，必须满足
	FeatureBusSchedule   Feature = "bus_schedule"    // 校车时刻表
	FeatureProgram       Feature = "program"         // 培养方案
	FeatureStudyProgress Feature = "study_progress"  // 学业进度
	FeatureElectricity   Feature = "electricity"     // 电费查询
	FeaturePayment       Feature = "payment"         // 校园卡查询
	FeatureMap           Feature = "map"             // 校园地图
)

var validFeatures = map[Feature]bool{
	FeatureTimetable:     true,
	FeatureGradeQuery:    true,
	FeatureGPACalc:       true,
	FeatureCourseSelect:  true,
	FeatureExamSchedule:  true,
	FeatureLogin:         true,
	FeatureBusSchedule:   true,
	FeatureProgram:       true,
	FeatureStudyProgress: true,
	FeatureElectricity:   true,
	FeaturePayment:       true,
	FeatureMap:           true,
}

var schoolCodeRe = regexp.MustCompile(`^[A-Z0-9_-]{1,20}$`)

func IsValidFeature(f Feature) bool {
	return validFeatures[f]
}

func IsValidSchoolCode(code string) bool {
	return schoolCodeRe.MatchString(code)
}

func IsValidURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	if u.User != nil {
		return false
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}
	if isPrivateIP(host) {
		return false
	}
	return true
}

var reservedCIDRs = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"::1/128",
	"fe80::/10",
}

var reservedNets []*net.IPNet

func init() {
	for _, cidr := range reservedCIDRs {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("invalid built-in CIDR: " + cidr)
		}
		reservedNets = append(reservedNets, block)
	}
}

func isPrivateIP(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, block := range reservedNets {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

type School struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Website   string    `json:"website"`
	Features  []Feature `json:"features"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSchoolRequest struct {
	Code     string    `json:"code" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	Website  string    `json:"website" binding:"required"`
	Features []Feature `json:"features" binding:"required"`
}

type UpdateSchoolRequest struct {
	Name     *string    `json:"name"`
	Website  *string    `json:"website"`
	Features *[]Feature `json:"features"`
	Enabled  *bool      `json:"enabled"`
}
