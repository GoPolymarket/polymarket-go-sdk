package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/data"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: 30 * time.Second,
	}

	cfg := polymarket.DefaultConfig()
	cfg.HTTPClient = httpClient
	client := polymarket.NewClient(
		polymarket.WithConfig(cfg),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userAddr := os.Getenv("POLYMARKET_USER_ADDRESS")
	if userAddr == "" {
		log.Fatal("请设置 POLYMARKET_USER_ADDRESS 环境变量")
	}

	user := common.HexToAddress(userAddr)

	updateInterval := 30 * time.Second
	if interval := os.Getenv("UPDATE_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			updateInterval = d
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	fmt.Println("=== 持仓跟踪器已启动 ===")
	fmt.Printf("监控地址: %s\n", userAddr)
	fmt.Printf("更新间隔: %v\n", updateInterval)
	fmt.Println("========================")

	printPositions(ctx, client, user)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			fmt.Println("\n收到退出信号，正在关闭...")
			return
		case <-ticker.C:
			printPositions(ctx, client, user)
		}
	}
}

func printPositions(ctx context.Context, client *polymarket.Client, user common.Address) {
	positions, err := client.Data.Positions(ctx, &data.PositionsRequest{
		User: user,
	})
	if err != nil {
		log.Printf("获取持仓失败: %v", err)
		return
	}

	if len(positions) == 0 {
		fmt.Println("\n当前无持仓")
		return
	}

	fmt.Println("\n========== 当前持仓 ==========")

	for _, p := range positions {
		fmt.Printf("市场: %s\n", p.Title)
		fmt.Printf("  结果: %s\n", p.Outcome)
		fmt.Printf("  数量: %s\n", p.Size.String())
		fmt.Printf("  平均价格: $%s\n", p.AvgPrice.String())
		fmt.Printf("  当前价格: $%s\n", p.CurPrice.String())
		fmt.Printf("  当前价值: $%s\n", p.CurrentValue.String())
		fmt.Printf("  未实现盈亏: $%s\n", p.CashPnl.String())
		fmt.Printf("  盈亏比例: %s%%\n", p.PercentPnl.String())
		fmt.Println()
	}
}
