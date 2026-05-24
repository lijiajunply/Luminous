package handler

import (
	"encoding/json"
	"net/http"

	"luminous/internal/model"

	"github.com/gin-gonic/gin"
)

type AppHandler struct{}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

func (h *AppHandler) GetTagModel(c *gin.Context) {
	apiUrl := "https://appapi.xauat.site/api/App/5f278ffc-5a70-4805-a6bf-0543040981a8/latest?channelId=9e1a198a-a0c2-4017-b492-f2d0e5bee437"

	resp, err := http.Get(apiUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接到外部API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "获取标签失败"})
		return
	}

	var rawData model.RawApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析数据失败"})
		return
	}

	// 转换逻辑，对应 C# 的 List<ReleaseInfo> 初始化
	result := []model.ReleaseInfo{
		{
			TagName: rawData.ReleaseId,
			Name:    rawData.ReleaseId,
			Body:    rawData.Context,
			Assets: []model.AssetInfo{
				{
					Name: func() string {
						if len(rawData.Softs) > 0 {
							return rawData.Softs[0].Name
						}
						return ""
					}(),
					BrowserDownloadUrl: func() string {
						if len(rawData.Softs) > 0 {
							return rawData.Softs[0].SoftUrl
						}
						return ""
					}(),
				},
			},
		},
	}

	c.JSON(http.StatusOK, result)
}
