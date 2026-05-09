package xauat

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"luminous/internal/util"
)

const (
	paymentBaseURL = "https://ydfwpt.xauat.edu.cn"
	oauthTokenURL  = "https://ydfwpt.xauat.edu.cn/berserker-auth/oauth/token"
	keyboardURL    = "https://ydfwpt.xauat.edu.cn/berserker-secure/keyboard?type=Standard&order=0&synAccessSource=h5"
	turnoverURL    = "http://ydfwpt.xauat.edu.cn/berserker-search/search/personal/turnover"
	balanceURL     = "https://ydfwpt.xauat.edu.cn/berserker-app/ykt/tsm/queryCard"
)

var paymentCache = util.NewCacheWithName("payment")

// PaymentItem 消费记录
type PaymentItem struct {
	TurnoverType string  `json:"turnover_type"`
	DatetimeStr  string  `json:"datetime_str"`
	Resume       string  `json:"resume"`
	Tranamt      float64 `json:"tranamt"`
}

// PaymentResponse 消费记录响应
type PaymentResponse struct {
	Records []PaymentItem `json:"records"`
	Total   float64       `json:"total"`
}

// GetPaymentToken 获取支付系统访问令牌，缓存 1 小时
func GetPaymentToken(cardNum string) (string, error) {
	cacheKey := "payment_token:" + cardNum
	val, err := paymentCache.GetOrSet(cacheKey, 1*time.Hour, func() (interface{}, error) {
		kb, err := fetchKeyboard()
		if err != nil {
			return "", fmt.Errorf("keyboard request failed: %w", err)
		}

		pwd := buildEncryptedPassword(kb)
		token, err := fetchOAuthToken(cardNum, pwd)
		if err != nil {
			return "", fmt.Errorf("oauth token request failed: %w", err)
		}

		return token, nil
	})
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// GetTurnover 获取消费记录与余额
func GetTurnover(cardNum string) (*PaymentResponse, error) {
	token, err := GetPaymentToken(cardNum)
	if err != nil {
		return nil, fmt.Errorf("payment login failed: %w", err)
	}

	var (
		records []PaymentItem
		balance float64
		mu      sync.Mutex
		wg      sync.WaitGroup
		errs    []error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		r, err := fetchTurnoverRecords(token, cardNum)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			records = r
		}
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		b, err := fetchBalance(token, cardNum)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			balance = b
		}
		mu.Unlock()
	}()

	wg.Wait()

	if len(records) == 0 && balance == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("get turnover failed: %w", errs[0])
	}
	for _, err := range errs {
		slog.Warn("partial payment fetch failed", "error", err)
	}

	return &PaymentResponse{Records: records, Total: balance}, nil
}

type keyboardResponse struct {
	UUID   string
	Digits []string
}

func fetchKeyboard() (*keyboardResponse, error) {
	resp, err := util.DefaultClient.Get(keyboardURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			NumberKeyboard []struct {
				Num string `json:"num"`
			} `json:"numberKeyboard"`
			UUID string `json:"uuid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse keyboard json: %w", err)
	}

	digits := make([]string, len(result.Data.NumberKeyboard))
	for i, n := range result.Data.NumberKeyboard {
		digits[i] = n.Num
	}

	return &keyboardResponse{
		UUID:   result.Data.UUID,
		Digits: digits,
	}, nil
}

func buildEncryptedPassword(kb *keyboardResponse) string {
	d := kb.Digits
	return fmt.Sprintf("%s%s%s%s%s%s$1%s", d[2], d[0], d[2], d[4], d[1], d[1], kb.UUID)
}

func fetchOAuthToken(cardNum, encryptedPwd string) (string, error) {
	resp, err := util.DefaultClient.PostForm(oauthTokenURL, url.Values{
		"username":        {cardNum},
		"grant_type":      {"password"},
		"scope":           {"all"},
		"loginFrom":       {"h5"},
		"logintype":       {"snoNew"},
		"device_token":    {"h5"},
		"synAccessSource": {"h5"},
		"password":        {encryptedPwd},
	}, map[string]string{
		"Authorization": paymentOAuthSecret,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse oauth token json: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}
	return result.AccessToken, nil
}

func fetchTurnoverRecords(token, cardNum string) ([]PaymentItem, error) {
	cacheKey := "payment_list:" + cardNum
	val, err := paymentCache.GetOrSet(cacheKey, 20*time.Minute, func() (interface{}, error) {
		resp, err := util.DefaultClient.GetWithHeaders(
			turnoverURL+"?size=8&current=1&synAccessSource=h5",
			map[string]string{
				"synAccessSource": "h5",
				"synjones-auth":   token,
			},
		)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result struct {
			Data struct {
				Records []struct {
					TurnoverType string  `json:"turnoverType"`
					DatetimeStr  string  `json:"jndatetimeStr"`
					Resume       string  `json:"resume"`
					Tranamt      float64 `json:"tranamt"`
				} `json:"records"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parse turnover json: %w", err)
		}

		records := make([]PaymentItem, 0, len(result.Data.Records))
		for _, r := range result.Data.Records {
			records = append(records, PaymentItem{
				TurnoverType: r.TurnoverType,
				DatetimeStr:  r.DatetimeStr,
				Resume:       r.Resume,
				Tranamt:      r.Tranamt / 100,
			})
		}
		return records, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]PaymentItem), nil
}

func fetchBalance(token, cardNum string) (float64, error) {
	cacheKey := "payment_balance:" + cardNum
	val, err := paymentCache.GetOrSet(cacheKey, 20*time.Minute, func() (interface{}, error) {
		resp, err := util.DefaultClient.GetWithHeaders(
			balanceURL+"?synAccessSource=h5",
			map[string]string{
				"synAccessSource": "h5",
				"synjones-auth":   token,
			},
		)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result struct {
			Data struct {
				Card []struct {
					ElecAccAmt float64 `json:"elec_accamt"`
				} `json:"card"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parse balance json: %w", err)
		}
		if len(result.Data.Card) == 0 {
			return float64(0), nil
		}
		return result.Data.Card[0].ElecAccAmt / 100, nil
	})
	if err != nil {
		return 0, err
	}
	return val.(float64), nil
}
