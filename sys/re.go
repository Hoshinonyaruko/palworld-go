package sys

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/net/html"
)

// Restarter is the interface that wraps the Restart method.
type Restarter interface {
	Restart(executableName string) error
}

// findIFrameSrc 遍历HTML节点以找到iframe标签的src属性
func findIFrameSrc(n *html.Node) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "iframe" {
		for _, a := range n.Attr {
			if a.Key == "src" {
				return a.Val, true
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if src, found := findIFrameSrc(c); found {
			return src, true
		}
	}
	return "", false
}

// GetExecutableName 返回当前执行文件的名称
func GetExecutableName() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(executable, filepath.Ext(executable)), nil
}

// linux
func setConsoleTitleLinux(title string) {
	fmt.Printf("\033]0;%s\007", title)
}

// windows
func setConsoleTitleWindows(title string) error {
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	proc, err := kernel32.FindProc("SetConsoleTitleW")
	if err != nil {
		return err
	}
	p0, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	r1, _, err := proc.Call(uintptr(unsafe.Pointer(p0)))
	if r1 == 0 {
		return err
	}
	return nil
}

// 判断系统
func setConsoleTitle(title string) error {
	switch runtime.GOOS {
	case "windows":
		return setConsoleTitleWindows(title)
	case "linux":
		setConsoleTitleLinux(title)
	default:
		fmt.Fprintf(os.Stderr, "setConsoleTitle not implemented for %s\n", runtime.GOOS)
	}
	return nil
}

// SetTitle sets the window title to "gensokyo-kook © 2023 - [Year] Hoshinonyaruko".
func SetTitle(title string) {
	err := setConsoleTitle(title)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set title: %v\n", err)
	}
}

// RestartApplication 封装了应用程序的重启逻辑
func RestartApplication() {
	execName, err := GetExecutableName() // 确保这个函数返回正确
	if err != nil {
		log.Println("Error getting executable name:", err)
		os.Exit(1) // 出错时退出码不为0
	}

	restarter := NewRestarter()
	if err := restarter.Restart(execName); err != nil {
		log.Println("Error restarting application:", err)
		os.Exit(1) // 出错时退出码不为0
	}

	// 创建restart.flag文件，表示自己正在restart
	if _, err := os.Create("restart.flag"); err != nil {
		log.Println("Unable to create restart flag:", err)
		os.Exit(1) // 出错时退出码不为0
	}

	// 退出程序
	os.Exit(0)
}

func findIP(node *html.Node) (string, bool) {
	if node.Type == html.ElementNode && node.Data == "iframe" {
		for _, a := range node.Attr {
			if a.Key == "src" {
				// We found the iframe, now let's send a request to the src URL
				resp, err := http.Get(a.Val)
				if err != nil {
					return "", false
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return "", false
				}
				doc, err := html.Parse(resp.Body)
				if err != nil {
					return "", false
				}
				// The actual IP might be in a <div> or <span>, depending on the page structure
				var f func(*html.Node) string
				f = func(n *html.Node) string {
					if n.Type == html.TextNode && strings.HasPrefix(n.Data, "您的IP地址信息:") {
						return strings.TrimSpace(strings.Split(n.Data, " ")[1])
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						if result := f(c); result != "" {
							return result
						}
					}
					return ""
				}
				ip := f(doc)
				if ip != "" {
					return ip, true
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if ip, found := findIP(c); found {
			return ip, true
		}
	}
	return "", false
}
