package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var username string           // 学号
var password string           // 密码
var operator string           // 运营商
var ssid string               // WiFi名称
var version string = "1.0.0"  // 版本号
var version_bool bool = false // 是否输出版本号

func init() {
	// 参数列表
	flag.StringVar(&username, "u", "", "学号")
	flag.StringVar(&password, "p", "", "密码")
	flag.StringVar(&operator, "o", "", "运营商：校园网，中国移动，中国电信，中国联通")
	flag.StringVar(&ssid, "w", "", "WiFi名称（可选，不填则跳过连接）")
	flag.BoolVar(&version_bool, "v", 1 == 2, "版本号")
	flag.Parse()
	if version_bool {
		fmt.Println(version)
		os.Exit(0)
	}

}
func main() {
	var logo = `                  ███  ████  █████                                        █████
                 ▒▒▒  ▒▒███ ▒▒███                                        ▒▒███
 █████████████   ████  ▒███  ▒███ █████  ██████   ██████   ████████    ███████  █████ ████
▒▒███▒▒███▒▒███ ▒▒███  ▒███  ▒███▒▒███  ███▒▒███ ▒▒▒▒▒███ ▒▒███▒▒███  ███▒▒███ ▒▒███ ▒███
 ▒███ ▒███ ▒███  ▒███  ▒███  ▒██████▒  ▒███ ▒▒▒   ███████  ▒███ ▒███ ▒███ ▒███  ▒███ ▒███
 ▒███ ▒███ ▒███  ▒███  ▒███  ▒███▒▒███ ▒███  ███ ███▒▒███  ▒███ ▒███ ▒███ ▒███  ▒███ ▒███
 █████▒███ █████ █████ █████ ████ █████▒▒██████ ▒▒████████ ████ █████▒▒████████ ▒▒███████
▒▒▒▒▒ ▒▒▒ ▒▒▒▒▒ ▒▒▒▒▒ ▒▒▒▒▒ ▒▒▒▒ ▒▒▒▒▒  ▒▒▒▒▒▒   ▒▒▒▒▒▒▒▒ ▒▒▒▒ ▒▒▒▒▒  ▒▒▒▒▒▒▒▒   ▒▒▒▒▒███
                                                                                 ███ ▒███
                                                                                ▒▒██████
                                                                                 ▒▒▒▒▒▒   `
	fmt.Println(logo)
	if username == "" || password == "" || operator == "" {
		fmt.Println("参数不足，请\"-h\"查看具体参数情况")
		return
	}

	// 连接WiFi
	if ssid != "" {
		if hasWiredConnection() {
			fmt.Println("检测到有线连接，跳过 WiFi 连接")
		} else {
			if !isConnected(ssid) {
				fmt.Println("正在连接 WiFi:", ssid)
				if err := connectWiFi(ssid); err != nil {
					fmt.Println("连接失败:", err)
					return
				}
				time.Sleep(3 * time.Second)
				if !isConnected(ssid) {
					fmt.Println("WiFi 连接超时，请检查 WiFi 名称或密码配置")
					return
				}
			}
			fmt.Println("WiFi 已连接:", ssid)
		}
	}

	fmt.Println("学号" + username)
	fmt.Println("密码" + password)
	fmt.Println("运营商" + operator)
	// 发送请求
	logourl := "http://10.11.0.10:801/eportal/portal/login?" +
		"callback=dr1006&login_method=1" +
		fmt.Sprintf("&user_account=%%2C0%%2C%s%%%s&user_password=%s&wlan_user_ip=%s", username, get_operator(), password, get_ip()) + // 学号，运营商，密码,本机ip
		"&wlan_user_ipv6=&wlan_user_mac=000000000000&wlan_ac_ip=&wlan_ac_name=&jsVersion=4.1.3&terminal_type=1&lang=zh-cn&v=3547&lang=zh"
	req, err := http.NewRequest("GET", logourl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 设置 referer，登录校园网需要检测来源
	req.Header.Set("Referer", "http://10.11.0.10/")

	// 模仿浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 关闭 body
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)
	//  读取body的内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(body))
}

// 获取本机ip
func get_ip() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if ipnet.IP.To4()[0] == 10 {
					ip := ipnet.IP.String()
					return ip
				}
			}
		} else {
			fmt.Println("ip不合法")
			continue
		}
	}
	fmt.Println("未获得有效ip")
	os.Exit(1)
	return ""
}

// 判断运营商
func get_operator() string {
	// 运营商对应代码
	switch operator {
	case "校园网":
		return ""
	case "中国移动":
		return "40cmcc"
	case "中国电信":
		return "40163.js"
	case "中国联通":
		return "40unicom"
	default:
		fmt.Println("运营商")
		os.Exit(1)
		return ""
	}
}

// 连接WiFi
func connectWiFi(ssid string) error {
	cmd := exec.Command("netsh", "wlan", "connect", "name="+ssid)
	return cmd.Run()
}

// 检查是否已连接指定WiFi
func isConnected(ssid string) bool {
	// 执行命令并获得输出
	out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), ssid)
}

// 断开WiFi
// func disconnectWiFi() error {
// 	return exec.Command("netsh", "wlan", "disconnect").Run()
// }

// 判断是否存在有线连接
func hasWiredConnection() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "wlan") ||
			strings.Contains(name, "wi-fi") ||
			strings.Contains(name, "wireless") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}

			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue
			}

			if ip4[0] == 10 { // 如果你只想认校园网地址
				return true
			}

			// 如果只要“有线已联网”就算，可以直接 return true
		}
	}

	return false
}
