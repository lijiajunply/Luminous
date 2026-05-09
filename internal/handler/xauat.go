package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"luminous/internal/response"
	"luminous/internal/school/xauat"

	"github.com/gin-gonic/gin"
)

type XAUATHandler struct{}

func NewXAUATHandler() *XAUATHandler {
	return &XAUATHandler{}
}

func (h *XAUATHandler) Code() string {
	return "XAUAT"
}

func (h *XAUATHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.GET("/courses", h.GetCourses)
	rg.GET("/scores", h.GetScores)
	rg.GET("/scores/semesters", h.GetSemesters)
	rg.GET("/scores/current-semester", h.GetCurrentSemester)
	rg.GET("/exams", h.GetExams)
	rg.GET("/bus", h.GetBus)
	rg.GET("/bus/:time", h.GetBus)
	rg.GET("/program", h.GetProgram)
	rg.GET("/info/completion", h.GetCompletion)
	rg.GET("/info/time", h.GetTimeInfo)
	rg.GET("/payment/:id", h.PaymentLogin)
	rg.GET("/payment/:id/turnover", h.PaymentTurnover)
}

// extractCookie 从请求中提取 Cookie（Cookie header 或 xauat header）
func extractCookie(c *gin.Context) string {
	cookie := c.GetHeader("Cookie")
	if cookie != "" && len(cookie) > 5 {
		return cookie
	}
	return c.GetHeader("xauat")
}

// Login SSO 登录
func (h *XAUATHandler) Login(c *gin.Context) {
	var req xauat.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: username and password required")
		return
	}

	result, err := xauat.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "login success", result)
}

// GetCourses 获取课表
func (h *XAUATHandler) GetCourses(c *gin.Context) {
	studentID := c.Query("student_id")
	if studentID == "" {
		response.Error(c, http.StatusBadRequest, "student_id is required")
		return
	}

	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	result, err := xauat.GetCourses(studentID, cookie)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetScores 获取成绩
func (h *XAUATHandler) GetScores(c *gin.Context) {
	studentID := c.Query("student_id")
	semester := c.Query("semester")

	if studentID == "" || semester == "" {
		response.Error(c, http.StatusBadRequest, "student_id and semester are required")
		return
	}

	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	scores, err := xauat.GetScores(studentID, semester, cookie)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", scores)
}

// GetSemesters 获取学期列表
func (h *XAUATHandler) GetSemesters(c *gin.Context) {
	studentID := c.Query("student_id")

	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	result, err := xauat.ParseSemesters(cookie, studentID)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetCurrentSemester 获取当前学期
func (h *XAUATHandler) GetCurrentSemester(c *gin.Context) {
	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	result, err := xauat.GetCurrentSemester(cookie)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetExams 获取考试安排
func (h *XAUATHandler) GetExams(c *gin.Context) {
	studentID := c.Query("student_id")
	if studentID == "" {
		response.Error(c, http.StatusBadRequest, "student_id is required")
		return
	}

	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	result, err := xauat.GetExams(cookie, studentID)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetBus 获取校车时刻表
func (h *XAUATHandler) GetBus(c *gin.Context) {
	date := c.Param("time")
	if date == "" {
		date = c.Query("date")
	}
	loc := c.Query("loc")
	if loc == "" {
		loc = "ALL"
	}

	result, err := xauat.GetBus(date, loc)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetProgram 获取培养方案
func (h *XAUATHandler) GetProgram(c *gin.Context) {
	programID := c.Query("id")
	if programID == "" {
		response.Error(c, http.StatusBadRequest, "id is required")
		return
	}
	filterName := c.Query("name")

	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	dict := c.Query("dict")
	if dict == "true" {
		result, err := xauat.GetProgramDict(programID, cookie)
		if err != nil {
			handleXAUATError(c, err)
			return
		}
		response.Success(c, http.StatusOK, "success", result)
		return
	}

	result, err := xauat.GetProgram(programID, cookie, filterName)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetCompletion 获取学业进度
func (h *XAUATHandler) GetCompletion(c *gin.Context) {
	cookie := extractCookie(c)
	if cookie == "" {
		response.Error(c, http.StatusUnauthorized, "cookie is required")
		return
	}

	result, err := xauat.GetCompletion(cookie)
	if err != nil {
		handleXAUATError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

// GetTimeInfo 获取学期时间范围
func (h *XAUATHandler) GetTimeInfo(c *gin.Context) {
	cookie := extractCookie(c)
	result := xauat.GetTimeInfo(cookie)
	response.Success(c, http.StatusOK, "success", result)
}

// PaymentLogin 获取支付系统令牌
func (h *XAUATHandler) PaymentLogin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "card number is required")
		return
	}

	token, err := xauat.GetPaymentToken(id)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "login success", token)
}

// PaymentTurnover 获取消费记录与余额
func (h *XAUATHandler) PaymentTurnover(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "card number is required")
		return
	}

	result, err := xauat.GetTurnover(id)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "success", result)
}

func handleXAUATError(c *gin.Context, err error) {
	var authErr *xauat.AuthError
	if errors.As(err, &authErr) {
		response.Error(c, http.StatusUnauthorized, authErr.Error())
		return
	}
	slog.Error("xauat internal error", "error", err)
	response.Error(c, http.StatusInternalServerError, "internal server error")
}
