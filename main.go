// main.go
package main

import (
	"fmt"
	"log"

	"os"
	"strconv"
	"strings"
	"time"

	"HuaTug.com/amap"
	"HuaTug.com/feishu"
	"gopkg.in/yaml.v3"
)

// Config 结构体用于映射 config.yaml
type Config struct {
	Amap struct {
		Key string `yaml:"key"`
	} `yaml:"amap"`
	Feishu struct {
		AppID     string `yaml:"app_id"`
		AppSecret string `yaml:"app_secret"`
		DocToken  string `yaml:"doc_token"`
	} `yaml:"feishu"`
}

// 加载配置
func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return &cfg, err
}

// 格式化路线信息
func formatRoute(origin, dest string, route *amap.DrivingResponse) string {
	if route == nil || len(route.Route.Paths) == 0 {
		return "未找到有效路线。"
	}

	path := route.Route.Paths[0]
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("路线查询【%s】\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("从 %s 到 %s\n", origin, dest))

	dist, _ := strconv.ParseFloat(path.Distance, 64)
	dura, _ := strconv.ParseFloat(path.Duration, 64)
	builder.WriteString(fmt.Sprintf("总距离: %.2f 公里\n", dist/1000))
	builder.WriteString(fmt.Sprintf("预计用时: %.0f 分钟\n", dura/60))
	builder.WriteString("--------------------\n")
	builder.WriteString("具体步骤:\n")

	for i, step := range path.Steps {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, step.Instruction))
	}

	builder.WriteString("\n\n") // 添加空行以便下次写入
	return builder.String()
}

func main() {
	// 1. 加载配置
	cfg, err := loadConfig("/Users/a1/Golang/mcp/config.yaml")
	if err != nil {
		log.Fatalf("无法加载配置文件 'config.yaml': %v", err)
	}

	// 2. 定义起点和终点
	originName:="深圳市南山区金地威新大厦"
	destinationName := "深圳市南山区滨海大厦"

	// 3. 初始化客户端
	amapClient := amap.NewClient(cfg.Amap.Key)
	feishuClient := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret)

	log.Printf("正在查询从 [%s] 到 [%s] 的路线...", originName, destinationName)

	origin,err:=amapClient.AddressToCoordinates(originName)
	if err != nil {
		log.Fatalf("转换地址失败: %v", err)
	}
	destination,err:=amapClient.AddressToCoordinates(destinationName)
	if err != nil {
		log.Fatalf("转换地址失败: %v", err)
	}
	// 4. 从高德获取路线
	routeData, err := amapClient.GetDrivingRoute(origin, destination)
	if err != nil {
		log.Fatalf("查询高德路线失败: %v", err)
	}
	log.Println("路线查询成功！")

	// 5. 格式化路线文本
	formattedText := formatRoute(origin, destination, routeData)

	// 6. 将文本写入飞书文档
	log.Printf("正在将路线写入飞书文档 (Token: %s)...", cfg.Feishu.DocToken)
	err = feishuClient.AppendTextToDoc(cfg.Feishu.DocToken, formattedText)
	if err != nil {
		log.Fatalf("写入飞书文档失败: %v", err)
	}

	log.Println("操作成功完成！路线已写入指定的飞书文档。")
}
