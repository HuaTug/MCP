// amap/client.go
package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	walkingURL   = "https://restapi.amap.com/v3/direction/walking" //步行路线
	GeocodingURL = "https://restapi.amap.com/v3/geocode/geo"       //地理编码
)

// GeocodingRequest 地理编码请求参数
type GeocodingRequest struct {
	Address    string // 待转换的地址
	City       string // 城市（可选，提高精度）
	Output     string // 返回格式（json/xml）
	Batch      string // 是否批量查询
	Sig        string // 签名（如需要）
	Extensions string // 返回结果详细程度
}

// GeocodingResponse 地理编码响应结构
type GeocodingResponse struct {
	Status   string    `json:"status"`   // 状态码（1成功）
	Info     string    `json:"info"`     // 状态说明
	Count    string    `json:"count"`    // 结果数量
	Geocodes []Geocode `json:"geocodes"` // 地理编码结果
}

// Geocode 单个地理编码结果
type Geocode struct {
	FormattedAddress string `json:"formatted_address"` // 格式化地址
	Province         string `json:"province"`          // 省份
	City             string `json:"city"`              // 城市
	District         string `json:"district"`          // 区县
	Location         string `json:"location"`          // 经纬度（经度,纬度）
	Level            string `json:"level"`             // 地址级别
}

// Amap 响应结构体 (根据高德API文档简化)
type DrivingResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Route  struct {
		Paths []struct {
			Distance string `json:"distance"`
			Duration string `json:"duration"`
			Steps    []struct {
				Instruction string `json:"instruction"`
			} `json:"steps"`
		} `json:"paths"`
	} `json:"route"`
}

// Client 是高德API的客户端
type Client struct {
	Key    string
	Client *http.Client
}


// NewClient 创建一个新的高德客户端
func NewClient(key string) *Client {
	return &Client{
		Key:    key,
		Client: &http.Client{},
	}
}

// GetDrivingRoute 查询驾车路线
func (c *Client) GetDrivingRoute(origin, destination string) (*DrivingResponse, error) {
	params := url.Values{}
	params.Add("key", c.Key)
	params.Add("origin", origin)
	params.Add("destination", destination)
	params.Add("extensions", "base") // 只获取基础信息

	reqURL := walkingURL + "?" + params.Encode()
	resp, err := c.Client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("请求高德API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取高德响应体失败: %w", err)
	}

	var drivingResp DrivingResponse
	if err := json.Unmarshal(body, &drivingResp); err != nil {
		return nil, fmt.Errorf("解析高德响应JSON失败: %w", err)
	}

	if drivingResp.Status != "1" {
		return nil, fmt.Errorf("高德API返回错误: %s", drivingResp.Info)
	}

	return &drivingResp, nil
}

func (c *Client)AddressToCoordinates(address string) (string, error) {
	// 构建请求参数
	params := url.Values{}
	params.Add("key", c.Key)
	params.Add("address", address)

	// 构建完整URL
	fullURL := fmt.Sprintf("%s?%s", GeocodingURL, params.Encode())

	// 发送GET请求
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("请求地理编码API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := readAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var geocodingResp GeocodingResponse
	if err := json.Unmarshal(body, &geocodingResp); err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API状态
	if geocodingResp.Status != "1" {
		return "", fmt.Errorf("地理编码失败: %s", geocodingResp.Info)
	}

	// 检查结果数量
	if len(geocodingResp.Geocodes) == 0 {
		return "", fmt.Errorf("未找到匹配的地址")
	}

	// 返回第一个结果的经纬度
	return geocodingResp.Geocodes[0].Location, nil
}

// 辅助函数：读取响应体并处理错误
func readAll(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return body, nil
}
