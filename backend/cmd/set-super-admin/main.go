package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❌ 错误: 请提供用户名")
		fmt.Println("")
		fmt.Println("使用方法:")
		fmt.Println("  set-super-admin <username>")
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  set-super-admin alice")
		os.Exit(1)
	}

	username := os.Args[1]

	// 从环境变量读取数据库配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ai_chat_user")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "ai_chat_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	if dbPassword == "" {
		fmt.Println("❌ 错误: DB_PASSWORD 环境变量未设置")
		os.Exit(1)
	}

	// 连接数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  设置超级管理员")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
	fmt.Printf("目标用户: %s\n", username)
	fmt.Println("")

	// 确认（非交互式环境直接执行）
	if os.Getenv("AUTO_CONFIRM") != "true" {
		fmt.Print("⚠️  确定要将用户设置为超级管理员吗？[y/N] ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("❌ 操作已取消")
			os.Exit(0)
		}
	}

	fmt.Println("")
	fmt.Println("正在设置超级管理员...")
	fmt.Println("")

	// 更新用户角色
	query := `
		UPDATE users
		SET role = 'super_admin',
		    updated_at = $1
		WHERE username = $2
		RETURNING id, username, role, created_at
	`

	var userID, role, createdAt string
	var returnedUsername string

	err = db.QueryRow(query, time.Now(), username).Scan(&userID, &returnedUsername, &role, &createdAt)
	if err == sql.ErrNoRows {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("❌ 失败！用户 '%s' 不存在\n", username)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("")
		fmt.Println("请先注册该用户（通过网页或OAuth2登录）")
		fmt.Println("")
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("❌ 失败！无法设置超级管理员: %v\n", err)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		os.Exit(1)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✅ 成功！用户 '%s' 已设置为超级管理员\n", username)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
	fmt.Println("用户信息:")
	fmt.Printf("  ID: %s\n", userID)
	fmt.Printf("  用户名: %s\n", returnedUsername)
	fmt.Printf("  角色: %s\n", role)
	fmt.Printf("  创建时间: %s\n", createdAt)
	fmt.Println("")
	fmt.Println("注意事项:")
	fmt.Println("  1. 超级管理员拥有所有权限")
	fmt.Println("  2. 可以管理其他管理员")
	fmt.Println("  3. 只能通过此工具设置，后台界面无法设置")
	fmt.Println("  4. 请妥善保管超级管理员账号")
	fmt.Println("")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
