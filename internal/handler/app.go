package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"luminous/internal/config"
	"luminous/internal/model"
	"luminous/internal/util"

	"github.com/gin-gonic/gin"
)

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

	resp, err := util.DefaultClient.Get(apiUrl)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream unreachable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream returned error"})
		return
	}

	var rawData model.RawApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid upstream response"})
		return
	}

	var name, url string
	if len(rawData.Softs) > 0 {
		name = rawData.Softs[0].Name
		url = rawData.Softs[0].SoftUrl
	}

	result := []model.ReleaseInfo{
		{
			TagName: rawData.ReleaseId,
			Name:    rawData.ReleaseId,
			Body:    rawData.Context,
			Assets: []model.AssetInfo{
				{Name: name, BrowserDownloadUrl: url},
			},
		},
	}

	c.JSON(http.StatusOK, result)
}
