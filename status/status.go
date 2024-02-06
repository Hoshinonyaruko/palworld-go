// status包
package status

var memoryIssueDetected bool = false
var successReadGameWorldSettings bool = false
var manualServerShutdown bool = false // 存储是否手动关闭服务器的状态
var globalpid int = 0

// SetMemoryIssueDetected 设置内存问题检测标志
func SetMemoryIssueDetected(flag bool) {
	memoryIssueDetected = flag
}

// GetMemoryIssueDetected 获取内存问题检测标志的当前值
func GetMemoryIssueDetected() bool {
	return memoryIssueDetected
}

// SetMemoryIssueDetected 设置内存问题检测标志
func SetsuccessReadGameWorldSettings(flag bool) {
	successReadGameWorldSettings = flag
}

// GetMemoryIssueDetected 获取内存问题检测标志的当前值
func GetsuccessReadGameWorldSettings() bool {
	return successReadGameWorldSettings
}

// SetManualServerShutdown 设置手动关闭服务器的状态
func SetManualServerShutdown(flag bool) {
	manualServerShutdown = flag
}

// GetManualServerShutdown 获取手动关闭服务器的状态
func GetManualServerShutdown() bool {
	return manualServerShutdown
}

func SetGlobalPid(pid int) {
	globalpid = pid
}

func GetGlobalPid() int {
	return globalpid
}
