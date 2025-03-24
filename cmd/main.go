package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"fsvchart-notify/internal/config"
	"fsvchart-notify/internal/database"
	"fsvchart-notify/internal/scheduler"
	"fsvchart-notify/internal/server"
	"fsvchart-notify/internal/service"
)

// version 将在编译时通过 -ldflags 注入
var version string

// getVersion 获取版本号，优先使用编译时注入的版本信息
func getVersion() string {
	// 如果编译时已注入版本信息，直接返回
	if version != "" {
		return version
	}

	// 尝试从VERSION文件读取版本号
	data, err := ioutil.ReadFile("VERSION")
	if err != nil {
		// 如果读取失败，返回默认版本
		return "unknown"
	}
	// 去除空白字符并返回
	return strings.TrimSpace(string(data))
}

func main() {
	// 设置 server 包的版本
	server.Version = getVersion()

	configPath := flag.String("config", "./config.yaml", "Path to config file")
	dbPath := flag.String("db", "./data/app.db", "Path to SQLite database file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	// 如果指定了 -version 参数，则显示版本号并退出
	if *showVersion {
		fmt.Printf("fsvchart-notify version %s\n", getVersion())
		os.Exit(0)
	}

	// 设置调试日志
	service.SetDebugLog(*debug)

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("LoadConfig error: %v", err)
	}

	// 初始化数据库
	_, err = database.InitDB(*dbPath)
	if err != nil {
		log.Fatalf("InitDB error: %v", err)
	}

	// 启动定时任务
	scheduler.StartScheduler()

	// 启动 Gin + Statik HTTP 服务
	srv := server.NewServer(cfg.Server.Address, cfg.Server.Port)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	log.Printf("fsvchart-notify %s running on %s", getVersion(), addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
