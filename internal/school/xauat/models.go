package xauat

import "time"

// --- Login ---

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Success   bool   `json:"success"`
	StudentID string `json:"student_id"`
	Cookie    string `json:"cookie"`
}

// --- Course ---

type CourseActivity struct {
	WeekIndexes []int    `json:"week_indexes"`
	Teachers    []string `json:"teachers"`
	Campus      string   `json:"campus"`
	Room        string   `json:"room"`
	CourseName  string   `json:"course_name"`
	CourseCode  string   `json:"course_code"`
	Weekday     int      `json:"weekday"`
	StartUnit   int      `json:"start_unit"`
	EndUnit     int      `json:"end_unit"`
	Credits     string   `json:"credits"`
	LessonID    string   `json:"lesson_id"`
}

type courseResponse struct {
	StudentTableVm struct {
		Activities []CourseActivity `json:"activities"`
	} `json:"studentTableVm"`
}

type CourseResultResponse struct {
	Success        bool              `json:"success"`
	Data           []CourseActivity  `json:"data"`
	ExpirationTime time.Time         `json:"expiration_time"`
}

// --- Score ---

type ScoreItem struct {
	Name        string `json:"name"`
	LessonCode  string `json:"lesson_code"`
	LessonName  string `json:"lesson_name"`
	Grade       string `json:"grade"`
	GPA         string `json:"gpa"`
	GradeDetail string `json:"grade_detail"`
	Credit      string `json:"credit"`
	IsMinor     bool   `json:"is_minor"`
}

// --- Semester ---

type SemesterItem struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

type SemesterResult struct {
	Data []SemesterItem `json:"data"`
}

// --- Exam ---

type ExamInfo struct {
	Name     string `json:"name"`
	Time     string `json:"time"`
	Location string `json:"location"`
	Seat     string `json:"seat"`
}

type ExamResponse struct {
	Exams    []ExamInfo `json:"exams"`
	CanClick bool       `json:"can_click"`
	Error    string     `json:"error,omitempty"`
}

type examDataRaw struct {
	Course   struct {
		NameZh string `json:"nameZh"`
	} `json:"course"`
	ExamTime string `json:"examTime"`
	Room     string `json:"room"`
	SeatNo   string `json:"seatNo"`
}

// --- Bus ---

type BusItem struct {
	LineName           string `json:"line_name"`
	Description        string `json:"description"`
	DepartureStation   string `json:"departure_station"`
	ArrivalStation     string `json:"arrival_station"`
	RunTime            string `json:"run_time"`
	ArrivalStationTime string `json:"arrival_station_time"`
}

type BusResponse struct {
	Records []BusItem `json:"records"`
	Total   int       `json:"total"`
}

// --- Program ---

type PlanCourse struct {
	Name           string  `json:"name"`
	LessonType     string  `json:"lesson_type"`
	ExamMode       string  `json:"exam_mode"`
	CourseTypeName string  `json:"course_type_name"`
	Credits        float64 `json:"credits"`
	TermStr        string  `json:"term_str"`
}

type programModule struct {
	Children    []programModule `json:"children"`
	PlanCourses []PlanCourse    `json:"planCourses"`
}

// --- Info ---

type CreditInfo struct {
	Name   string  `json:"name"`
	Actual float64 `json:"actual"`
	Full   float64 `json:"full"`
}

type StudyModule struct {
	Type  string       `json:"type"`
	Total CreditInfo   `json:"total"`
	Other []CreditInfo `json:"other"`
}

