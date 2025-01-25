package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"math/rand/v2" // 使用 v2 版本的 rand 包
	"net/http"
	"sync"
	"time"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	// 启动定时任务，每分钟清理一次过期的代码片段
	go cleanupExpiredSnippets()

	// 定义一个命令行参数
	var route string
	flag.StringVar(&route, "route", "", "The route to use for sharing code snippets")
	flag.Parse()

	// 如果没有提供路由参数，则生成一个随机的 ID
	if route == "" {
		route = generateSnippetID()
	}

	http.HandleFunc("/"+route, shareHandler)
	http.HandleFunc("/view/", viewHandler)

	fmt.Println("Starting server at http://127.0.0.1:8003/" + route)
	err := http.ListenAndServe(":8003", nil)
	if err != nil {
		fmt.Println("Starting server error", err)
		return
	}
}

var (
	snippets = make(map[string]CodeSnippet)
	mu       sync.Mutex // 用于保护 snippets 的并发访问
)

type CodeSnippet struct {
	Code      string
	CodeType  string
	ExpiresAt time.Time
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		codeType := r.FormValue("codeType") // 获取代码类型
		expires := r.FormValue("expires")
		code := r.FormValue("code")

		fmt.Println("shareHandler: CodeType:", codeType, "Expires:", expires)

		expiresDuration, _ := time.ParseDuration(expires)
		expiresAt := time.Now().Add(expiresDuration)

		snippetID := generateSnippetID()
		mu.Lock()
		snippets[snippetID] = CodeSnippet{
			Code:      code,
			CodeType:  codeType,
			ExpiresAt: expiresAt,
		}
		mu.Unlock()

		url := "/view/" + snippetID
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	// 解析嵌入的模板文件
	shareTmpl, err := template.ParseFS(templatesFS, "templates/share.html")
	if err != nil {
		fmt.Println(err)
	}
	err = shareTmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	snippetID := r.URL.Path[len("/view/"):]
	mu.Lock()
	snippet, exists := snippets[snippetID]
	mu.Unlock()

	// 检查代码片段是否过期
	if !exists || time.Now().After(snippet.ExpiresAt) {
		// 如果过期，立即删除
		mu.Lock()
		delete(snippets, snippetID)
		mu.Unlock()
		http.Error(w, "Snippet not found or expired", http.StatusNotFound)
		return
	}

	// 将过期时间转换为北京时间
	location, _ := time.LoadLocation("Asia/Shanghai")
	expiresAtBeijing := snippet.ExpiresAt.In(location)

	data := struct {
		Code      string
		CodeType  string
		ExpiresAt string
	}{
		Code:      snippet.Code,
		CodeType:  snippet.CodeType,
		ExpiresAt: expiresAtBeijing.Format("2006-01-02 15:04:05"),
	}

	// 解析嵌入的模板文件
	viewTmpl, err := template.ParseFS(templatesFS, "templates/view.html")
	if err != nil {
		fmt.Println(err)
	}
	err = viewTmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength    = 8
)

func generateSnippetID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))] // 使用 rand.IntN
	}
	return string(b)
}

// cleanupExpiredSnippets 每分钟清理一次过期的代码片段
func cleanupExpiredSnippets() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()
		now := time.Now()
		for id, snippet := range snippets {
			if now.After(snippet.ExpiresAt) {
				fmt.Printf("Deleting expired snippet: %s ExpiresAt: %s\n", id, snippet.ExpiresAt)
				delete(snippets, id)
			}
		}
		mu.Unlock()
	}
}
