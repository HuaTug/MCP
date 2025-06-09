package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	tokenURL       = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	appendBlockURL = "https://open.feishu.cn/open-apis/docx/v1/documents/%s/blocks/%s/children"
)

// Client 是飞书API的客户端
type Client struct {
	AppID     string
	AppSecret string
	Client    *http.Client
	token     *tenantAccessToken
}

// tenantAccessToken 结构体
type tenantAccessToken struct {
	Token     string    `json:"tenant_access_token"`
	Expire    int       `json:"expire"`
	ExpiresAt time.Time // 用于判断Token是否过期
}

// NewClient 创建一个新的飞书客户端
func NewClient(appID, appSecret string) *Client {
	return &Client{
		AppID:     appID,
		AppSecret: appSecret,
		Client:    &http.Client{},
	}
}

// 获取或刷新 tenant_access_token
func (c *Client) refreshToken() (string, error) {
	// 如果token存在且未过期，直接返回
	if c.token != nil && time.Now().Before(c.token.ExpiresAt) {
		return c.token.Token, nil
	}

	reqBody := map[string]string{
		"app_id":     c.AppID,
		"app_secret": c.AppSecret,
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := c.Client.Post(tokenURL, "application/json; charset=utf-8", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("请求飞书Token API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取飞书Token响应体失败: %w", err)
	}

    var rawResp map[string]interface{}
    if err := json.Unmarshal(body, &rawResp); err != nil {
        return "", fmt.Errorf("解析飞书Token JSON失败: %w", err)
    }

    if code, ok := rawResp["code"].(float64); !ok || int(code) != 0 {
         return "", fmt.Errorf("获取飞书Token失败: %s", rawResp["msg"])
    }

	c.token = &tenantAccessToken{
		Token:  rawResp["tenant_access_token"].(string),
		Expire: int(rawResp["expire"].(float64)),
	}

	// 减去一分钟作为缓冲，防止边缘情况
	c.token.ExpiresAt = time.Now().Add(time.Duration(c.token.Expire-60) * time.Second)

	return c.token.Token, nil
}

// AppendTextToDoc 向飞书文档末尾追加文本块
func (c *Client) AppendTextToDoc(docToken, text string) error {
	token, err := c.refreshToken()
	if err != nil {
		return err
	}
	fmt.Println(text)
	
	// 使用"end"作为block_id，表示追加到文档末尾
    //const lastBlockID = "end"

	// 构造正确的请求体结构
	reqBody := map[string]interface{}{
		"children": []map[string]interface{}{
			{
				"block_type": 2, // 2代表文本块
				"text": map[string]interface{}{
					"elements": []map[string]interface{}{
						{
							"text_run": map[string]interface{}{
								"content": text,
							},
						},
					},
				},
			},
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	url := fmt.Sprintf(appendBlockURL, docToken, docToken)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("请求飞书Append API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("飞书Append API返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
