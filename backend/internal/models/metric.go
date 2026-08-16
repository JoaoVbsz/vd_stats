package models

import "time"

type ContainerMetric struct {
	DockerID   string    `json:"docker_id"`
	Name       string    `json:"name"`
	ProjectDir string    `json:"project_dir"`
	CPUPercent float64   `json:"cpu_percent"`
	MemUsed    int64     `json:"mem_used"`
	MemLimit   int64     `json:"mem_limit"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

type VPSMetric struct {
	Host       string    `json:"host"`
	CPUPercent float64   `json:"cpu_percent"`
	MemUsed    int64     `json:"mem_used"`
	MemTotal   int64     `json:"mem_total"`
	Timestamp  time.Time `json:"timestamp"`
}
