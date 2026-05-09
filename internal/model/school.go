package model

import "time"

type Feature string

const (
	FeatureTimetable     Feature = "timetable"        // 课表显示
	FeatureGradeQuery    Feature = "grade_query"      // 成绩查询
	FeatureGPACalc       Feature = "gpa_calculation"  // GPA计算
	FeatureCourseSelect  Feature = "course_selection" // 选课
	FeatureExamSchedule  Feature = "exam_schedule"    // 考试安排
	FeatureLogin         Feature = "login"            // SSO登录
	FeatureBusSchedule   Feature = "bus_schedule"     // 校车时刻表
	FeatureProgram       Feature = "program"          // 培养方案
	FeatureStudyProgress Feature = "study_progress"   // 学业进度
	FeatureSemesterInfo  Feature = "semester_info"    // 学期信息
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

func IsValidFeature(f Feature) bool {
	return validFeatures[f]
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
