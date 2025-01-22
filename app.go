package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// App struct
type App struct {
	ctx        context.Context
	mu         sync.Mutex
	progress   int
	cancelFunc context.CancelFunc // 用于取消检查
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Result 结构体，用于存储 URL 和文件大小
type Result struct {
	URL  string
	Size string
}

// CheckFileSizeConcurrent 并发检查 URL 文件大小
func (a *App) CheckFileSizeConcurrent(urls []string, concurrency int, outputFile string) ([]Result, error) {
	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel // 保存取消函数
	defer cancel()        // 确保检查完成后释放资源

	var wg sync.WaitGroup
	results := make([]Result, len(urls))
	queue := make(chan int, concurrency) // 控制并发数

	// 创建 HTTP 客户端，设置超时时间
	client := &http.Client{Timeout: 10 * time.Second}

	for i, url := range urls {
		select {
		case <-ctx.Done(): // 监听取消信号
			return nil, ctx.Err()
		default:
			wg.Add(1)
			queue <- i // 占用一个并发槽
			go func(index int, u string) {
				defer wg.Done()
				defer func() { <-queue }() // 释放并发槽

				size, err := getFileSize(ctx, client, u) // 传递 context 和 client
				if err != nil {
					results[index] = Result{URL: u, Size: "获取失败"}
				} else {
					results[index] = Result{URL: u, Size: formatFileSize(size)}
				}

				// 更新进度
				a.mu.Lock()
				a.progress = (index + 1) * 100 / len(urls)
				runtime.EventsEmit(a.ctx, "progress", a.progress)
				a.mu.Unlock()
			}(i, url)
		}
	}

	wg.Wait()

	// 按文件大小倒序排序
	sort.Slice(results, func(i, j int) bool {
		sizeI := parseSize(results[i].Size)
		sizeJ := parseSize(results[j].Size)
		return sizeI > sizeJ
	})

	// 写入 Excel 文件
	if err := writeToExcel(results, outputFile); err != nil {
		return nil, err
	}

	return results, nil
}

// getFileSize 获取指定 URL 文件的大小，支持 context 取消
func getFileSize(ctx context.Context, client *http.Client, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size <= 0 {
		return 0, errors.New("无法确定文件大小")
	}

	return size, nil
}

// formatFileSize 格式化文件大小为易读的字符串
func formatFileSize(size int64) string {
	switch {
	case size >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(size)/(1<<30))
	case size >= 1<<20:
		return fmt.Sprintf("%.2f MB", float64(size)/(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%.2f KB", float64(size)/(1<<10))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// parseSize 将格式化后的文件大小字符串解析为字节数
func parseSize(sizeStr string) int64 {
	if sizeStr == "获取失败" {
		return -1
	}
	var size float64
	var unit string
	fmt.Sscanf(sizeStr, "%f %s", &size, &unit)

	switch unit {
	case "GB":
		return int64(size * (1 << 30))
	case "MB":
		return int64(size * (1 << 20))
	case "KB":
		return int64(size * (1 << 10))
	case "B":
		return int64(size)
	default:
		return 0
	}
}

// writeToExcel 将结果写入 Excel 文件
func writeToExcel(results []Result, outputFile string) error {
	excel := excelize.NewFile()
	sheetName := "Results"
	excel.SetSheetName(excel.GetSheetName(0), sheetName)
	excel.SetCellValue(sheetName, "A1", "URL")
	excel.SetCellValue(sheetName, "B1", "文件大小")

	for i, result := range results {
		row := i + 2
		excel.SetCellValue(sheetName, fmt.Sprintf("A%d", row), result.URL)
		excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), result.Size)
	}

	if err := excel.SaveAs(outputFile); err != nil {
		return err
	}

	return nil
}

// CancelCheck 取消检查
func (a *App) CancelCheck() {
	if a.cancelFunc != nil {
		a.cancelFunc() // 调用取消函数
	}
}