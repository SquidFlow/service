package autopilot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Application 结构体定义
type Application struct {
	AppName     string `json:"app-name"`
	Repo        string `json:"repo"`
	WaitTimeout string `json:"wait-timeout"`
}

// WriteApplicationToFile 将应用程序数据写入文件
func WriteApplicationToFile(app Application) error {
	// 创建一个以应用名称和时间戳命名的文件
	fileName := fmt.Sprintf("%s_%d.json", app.AppName, time.Now().Unix())
	filePath := filepath.Join("applications", fileName)

	// 确保 applications 目录存在
	if err := os.MkdirAll("applications", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}

	// 将应用程序数据转换为 JSON
	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal application data: %w", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write application data to file: %w", err)
	}

	log.Printf("Application data saved to %s", filePath)
	return nil
}
