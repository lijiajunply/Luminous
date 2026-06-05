package model

import (
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
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
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
