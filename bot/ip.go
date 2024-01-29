package bot

import (
	"fmt"
	"strconv"
	"strings"
)

func ipToNumber(ip string) (int64, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("invalid IP address format")
	}

	r, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	g, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	b, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, err
	}

	c := int64(r) + int64(g)*256 + int64(b)*256*256 + int64(n)*256*256*256
	c = (c + 250) * 2
	return c, nil
}

func numberToIP(num int64) string {
	// 因为在转换时执行了 (c + 250) * 2
	// 所以先进行反向操作
	num = (num / 2) - 250

	// 提取IP的四个部分
	n := num / (256 * 256 * 256) % 256
	b := num / (256 * 256) % 256
	g := num / 256 % 256
	r := num % 256

	// 构造IP地址
	return fmt.Sprintf("%d.%d.%d.%d", r, g, b, n)
}

func IpToNumberWithPort(ip string) (int64, error) {
	parts := strings.Split(ip, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid IP address and port format")
	}

	ipPart, portPart := parts[0], parts[1]
	port, err := strconv.Atoi(portPart)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %s", err)
	}

	num, err := ipToNumber(ipPart)
	if err != nil {
		return 0, err
	}

	return num*65536 + int64(port), nil // 65536 = 256*256，确保端口不会与IP地址冲突
}

func numberToIPWithPort(num int64) string {
	port := num % 65536
	num = num / 65536

	ip := numberToIP(num)
	return fmt.Sprintf("%s:%d", ip, port)
}
