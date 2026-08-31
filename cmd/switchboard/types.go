package main

type ServiceState struct {
	Service
	Status    string  `json:"status"`
	Autostart bool    `json:"autostart"`
	CPU       float64 `json:"cpuPercent"`
	Memory    uint64  `json:"memoryBytes"`
	Error     string  `json:"error,omitempty"`
}

type SystemStats struct {
	CPUPercent    float64 `json:"cpuPercent"`
	CPUCores      int     `json:"cpuCores"`
	MemoryUsed    uint64  `json:"memoryUsedBytes"`
	MemoryFree    uint64  `json:"memoryFreeBytes"`
	MemoryTotal   uint64  `json:"memoryTotalBytes"`
	SwapUsed      uint64  `json:"swapUsedBytes"`
	SwapFree      uint64  `json:"swapFreeBytes"`
	SwapTotal     uint64  `json:"swapTotalBytes"`
	DiskUsed      uint64  `json:"diskUsedBytes"`
	DiskFree      uint64  `json:"diskFreeBytes"`
	DiskTotal     uint64  `json:"diskTotalBytes"`
	Temperature   float64 `json:"temperatureCelsius"`
	LoadOne       float64 `json:"loadOne"`
	UptimeSeconds uint64  `json:"uptimeSeconds"`
	Hostname      string  `json:"hostname"`
	LocalIP       string  `json:"localIp"`
	OS            string  `json:"os"`
	Kernel        string  `json:"kernel"`
	Architecture  string  `json:"architecture"`
}

type serviceBackend interface {
	State(Service) (ServiceState, error)
	Action(Service, string) error
	SetAutostart(Service, bool) error
}

type systemCollector interface {
	Stats() (SystemStats, error)
}
