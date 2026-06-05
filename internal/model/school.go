package model

import (
	"net/url"
	"regexp"
	"time"
)

type Feature string

const (
	FeatureTimetable     Feature = "timetable"
	FeatureGradeQuery    Feature = "grade_query"
	FeatureGPACalc       Feature = "gpa_calculation"
	FeatureCourseSelect  Feature = "course_selection"
	FeatureExamSchedule  Feature = "exam_schedule"
	FeatureLogin         Feature = "login"
	FeatureBusSchedule   Feature = "bus_schedule"
	FeatureProgram       Feature = "program"
	FeatureStudyProgress Feature = "study_progress"
	FeatureSemesterInfo  Feature = "semester_info"
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
	FeatureSemesterInfo:  true,
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
