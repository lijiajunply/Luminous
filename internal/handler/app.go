package handler

import (
	"encoding/json"
	"fmt"
	"io"
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

type AppHandler struct{}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

func (h *AppHandler) GetTagModel(c *gin.Context) {
	apiUrl := config.Cfg.Release.APIURL
	if apiUrl == "" {
		apiUrl = fmt.Sprintf(
			"https://appapi.xauat.site/api/App/%s/latest?channelId=%s",
			config.Cfg.Release.AppUUID,
			config.Cfg.Release.ChannelID,
		)
	}

	parsed, err := url.Parse(apiUrl)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "invalid upstream URL configured")
		return
	}
	if !allowedUpstreamHosts[parsed.Host] && !allowedUpstreamHosts[strings.SplitN(parsed.Host, ":", 2)[0]] {
		response.Error(c, http.StatusInternalServerError, "upstream host not allowed")
		return
	}

	resp, err := util.DefaultClient.Get(apiUrl)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "upstream unreachable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		response.Error(c, http.StatusBadGateway, "upstream returned error")
		return
	}

	var rawData model.RawApiResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rawData); err != nil {
		response.Error(c, http.StatusInternalServerError, "invalid upstream response")
		return
	}

	var name, downloadUrl string
	if len(rawData.Softs) > 0 {
		name = rawData.Softs[0].Name
		downloadUrl = rawData.Softs[0].SoftUrl
	}

	result := []model.ReleaseInfo{
		{
			TagName: rawData.ReleaseId,
			Name:    rawData.ReleaseId,
			Body:    rawData.Context,
			Assets: []model.AssetInfo{
				{Name: name, BrowserDownloadUrl: downloadUrl},
			},
		},
	}

	response.SuccessList(c, http.StatusOK, "success", len(result), result)
}
