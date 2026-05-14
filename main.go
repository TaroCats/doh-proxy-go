/*
 * @Author: taro etsy@live.com
 * @LastEditors: taro etsy@live.com
 * @LastEditTime: 2026-05-14 09:20:01
 * @Description: 一个简单的 DoH 代理，用于将加密的 DNS 查询报文转发给传统 DNS 服务器
 */
package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/miekg/dns"
)

// 你的传统 DNS 服务器地址
var upstreamDNS = getEnv("UPSTREAM_DNS", "124.221.68.73:1053")

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func dohHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取客户端发来的加密 DNS 查询报文 (遵循 RFC 8484 GET 请求规范)
	dnsParam := r.URL.Query().Get("dns")
	if dnsParam == "" {
		http.Error(w, "缺少 dns 参数", http.StatusBadRequest)
		return
	}

	// 2. Base64 解码还原为标准的 DNS 二进制报文
	dnsMsgBytes, err := base64.RawURLEncoding.DecodeString(dnsParam)
	if err != nil {
		http.Error(w, "DNS 报文解码失败", http.StatusBadRequest)
		return
	}

	// 3. 将 DNS 报文转发给你的传统 DNS (124.221.68.73:1053)
	dnsClient := new(dns.Client)
	dnsRequest := new(dns.Msg)
	if err := dnsRequest.Unpack(dnsMsgBytes); err != nil {
		http.Error(w, "DNS 报文解析失败", http.StatusBadRequest)
		return
	}

	// 通过 UDP 协议与你的传统 DNS 通信
	dnsResponse, _, err := dnsClient.Exchange(dnsRequest, upstreamDNS)
	if err != nil {
		log.Printf("转发至 %s 失败: %v", upstreamDNS, err)
		http.Error(w, "上游 DNS 查询失败", http.StatusInternalServerError)
		return
	}

	// 4. 将上游返回的结果打包并加密返回给客户端
	responseBytes, err := dnsResponse.Pack()
	if err != nil {
		http.Error(w, "响应打包失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(responseBytes)
}

func main() {
	// 本地监听 8053 端口提供 DoH 服务
	http.HandleFunc("/dns-query", dohHandler)
	fmt.Println("DoH 代理已启动，正在监听 :8053，上游指向:", upstreamDNS)
	log.Fatal(http.ListenAndServe(":8053", nil))
}
