package xauat

import "luminous/internal/config"

const defaultOAuthSecret = "Basic bW9iaWxlX3NlcnZpY2VfcGxhdGZvcm06bW9iaWxlX3NlcnZpY2VfcGxhdGZvcm1fc2VjcmV0"

var (
	baseURL            = "https://swjw.xauat.edu.cn"
	loginURL           = "https://schedule.xauat.site"
	oldBusURL          = "https://school-bus.xauat.edu.cn"
	newBusURL          = "https://bcdd.xauat.edu.cn"
	semesterStart      = "2026-03-01"
	semesterEnd        = "2026-07-18"
	paymentOAuthSecret = defaultOAuthSecret
)

// Init configures the XAUAT package from application config.
func Init(cfg config.XAUATConfig) {
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}
	if cfg.LoginURL != "" {
		loginURL = cfg.LoginURL
	}
	if cfg.OldBusURL != "" {
		oldBusURL = cfg.OldBusURL
	}
	if cfg.NewBusURL != "" {
		newBusURL = cfg.NewBusURL
	}
	if cfg.SemesterStart != "" {
		semesterStart = cfg.SemesterStart
	}
	if cfg.SemesterEnd != "" {
		semesterEnd = cfg.SemesterEnd
	}
	if cfg.PaymentOAuthSecret != "" {
		paymentOAuthSecret = cfg.PaymentOAuthSecret
	}
}
