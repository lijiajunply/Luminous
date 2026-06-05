package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"luminous/internal/config"
	"luminous/internal/model"
	"luminous/internal/response"
	"luminous/internal/util"

	"github.com/gin-gonic/gin"
)

var allowedUpstreamHosts = map[string]bool{
	"appapi.xauat.site": true,
}

type AppHandler struct {
	releaseCfg config.ReleaseConfig
	client     *util.HTTPClient
}

func NewAppHandler(cfg config.ReleaseConfig) *AppHandler {
	return &AppHandler{
		releaseCfg: cfg,
		client:     util.NewHTTPClientWithAllowedHosts(allowedUpstreamHosts),
	}
}

func (h *AppHandler) GetTagModel(c *gin.Context) {
	apiUrl := h.releaseCfg.APIURL
	if apiUrl == "" {
		apiUrl = fmt.Sprintf(
			"https://appapi.xauat.site/api/App/%s/latest?channelId=%s",
			h.releaseCfg.AppUUID,
			h.releaseCfg.ChannelID,
		)
	}

	parsed, err := url.Parse(apiUrl)
	if err != nil {
		slog.Error("invalid upstream URL", "request_id", c.GetString("request_id"), "error", err)
		response.Error(c, http.StatusInternalServerError, "invalid upstream URL configured")
		return
	}
	if !allowedUpstreamHosts[parsed.Host] && !allowedUpstreamHosts[strings.SplitN(parsed.Host, ":", 2)[0]] {
		slog.Error("upstream host not allowed", "request_id", c.GetString("request_id"), "host", parsed.Host)
		response.Error(c, http.StatusInternalServerError, "upstream host not allowed")
		return
	}

	resp, err := h.client.GetWithContext(c.Request.Context(), apiUrl)
	if err != nil {
		slog.Error("upstream unreachable", "request_id", c.GetString("request_id"), "error", err)
		response.Error(c, http.StatusBadGateway, "upstream unreachable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		slog.Error("upstream returned error", "request_id", c.GetString("request_id"), "status", resp.StatusCode)
		response.Error(c, http.StatusBadGateway, "upstream returned error")
		return
	}

	var rawData model.RawApiResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rawData); err != nil {
		slog.Error("invalid upstream response", "request_id", c.GetString("request_id"), "error", err)
		response.Error(c, http.StatusInternalServerError, "invalid upstream response")
		return
	}

	result := []model.ReleaseInfo{
		{
			TagName: rawData.ReleaseId,
			Name:    rawData.ReleaseId,
			Body:    rawData.Context,
			Assets:  nil,
		},
	}

	if len(rawData.Softs) > 0 {
		result[0].Assets = []model.AssetInfo{
			{Name: rawData.Softs[0].Name, BrowserDownloadUrl: rawData.Softs[0].SoftUrl},
		}
	}

	response.SuccessList(c, http.StatusOK, "success", len(result), result)
}
