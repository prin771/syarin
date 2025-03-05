package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
)

func parseArgs() string {
	filePath := flag.String("f", "", "IPアドレスを読み込むファイルのパス (必須)")
	help := flag.Bool("help", false, "ヘルプメッセージを表示")
	flag.Parse()
	if *help || *filePath == "" {
		flag.PrintDefaults()
		return ""
	}
	return *filePath
}

func compareIP(ip1, ip2 net.IP) bool {
	if len(ip1) == 0 && len(ip2) == 0 {
		return false
	}
	if len(ip1) == 0 {
		return true
	}
	if len(ip2) == 0 {
		return false
	}
	return string(ip1) < string(ip2)
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.To4() != nil {
		return ip.IsPrivate() ||
			ip.Equal(net.IPv4(127, 0, 0, 1))
	}
	return ip.IsPrivate()
}

func processIP(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("os.Open: %v\n", err)
		return
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("io.ReadAll: %v\n", err)
		return
	}

	ipRegex := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	ips := ipRegex.FindAllString(string(fileContent), -1)

	uniqueIPs := make(map[string]bool)
	for _, ip := range ips {
		if ip != "0.0.0.0" {
			uniqueIPs[ip] = true
		}
	}

	sortedIPs := make([]string, 0, len(uniqueIPs))
	for ip := range uniqueIPs {
		sortedIPs = append(sortedIPs, ip)
	}

	sort.Slice(sortedIPs, func(i, j int) bool {
		ip1 := net.ParseIP(sortedIPs[i])
		ip2 := net.ParseIP(sortedIPs[j])
		return ip1.To4() != nil && ip2.To4() == nil || compareIP(ip1, ip2)
	})

	if len(sortedIPs) == 0 {
		fmt.Println("ファイルにIPアドレスが見つかりませんでした。")
		return
	}

	for _, ipStr := range sortedIPs {
		ip := net.ParseIP(ipStr)
		if !isPrivateIP(ip) {
			domains, err := net.LookupAddr(ipStr)
			if err != nil {
				fmt.Printf("%s: 逆引き失敗 (%v)\n", ipStr, err)
			} else {
				fmt.Printf("%s: %s\n", ipStr, strings.Join(domains, ", "))
			}
		} else {
			fmt.Printf("%s: ローカルIP\n", ipStr)
		}
	}
}

func main() {
	filePath := parseArgs()
	if filePath != "" {
		processIP(filePath)
	}
}
