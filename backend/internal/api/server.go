package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/joaov/vd_stats/internal/database"
)

type ContainerLiveStat struct {
	ServerID string  `json:"server_id"`
	DockerID string  `json:"docker_id"`
	Name     string  `json:"name"`
	CPU      float64 `json:"cpu"`
	MemUsed  int64   `json:"mem_used"`
	MemLimit int64   `json:"mem_limit"`
}

type ServerLiveStat struct {
	ID            string  `json:"id"`
	HostIP        string  `json:"host_ip"`
	Name          string  `json:"name"`
	Uptime        float64 `json:"uptime"`
	DiskUsed      int64   `json:"disk_used"`
	DiskTotal     int64   `json:"disk_total"`
	PingLatencyMs float64 `json:"latency_ms"`
}

type LbStat struct {
	UpstreamAddr  string `json:"upstream_addr"`
	ServerName    string `json:"server_name"`
	Status        string `json:"status"`
	RequestsCount int    `json:"requests_count"`
}

type LiveResponse struct {
	Servers       []ServerLiveStat    `json:"servers"`
	Containers    []ContainerLiveStat `json:"containers"`
	LoadBalancing []LbStat            `json:"load_balancing"`
}

func StartServer(port string) {
	mux := http.NewServeMux()

	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Content-Type", "application/json")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			h(w, r)
		}
	}

	mux.HandleFunc("/api/metrics/live", withCORS(func(w http.ResponseWriter, r *http.Request) {
		var servers []database.Server
		database.DB.Find(&servers)

		var res LiveResponse
		
		for _, s := range servers {
			var lastSrv database.MetricServer
			database.DB.Where("server_id = ?", s.ID).Order("timestamp desc").First(&lastSrv)
			
			res.Servers = append(res.Servers, ServerLiveStat{
				ID:            s.ID,
				HostIP:        s.HostIP,
				Name:          s.Name,
				Uptime:        lastSrv.UptimeSeconds,
				DiskUsed:      lastSrv.DiskUsedBytes,
				DiskTotal:     lastSrv.DiskTotalBytes,
				PingLatencyMs: lastSrv.PingLatencyMs,
			})
		}

		var containers []database.Container
		database.DB.Find(&containers)

		for _, c := range containers {
			var lastMetric database.MetricContainer
			dbRes := database.DB.Where("container_id = ?", c.ID).Order("timestamp desc").First(&lastMetric)
			
			if dbRes.Error == nil {
				res.Containers = append(res.Containers, ContainerLiveStat{
					ServerID: c.ServerID,
					DockerID: c.DockerID,
					Name:     c.Name,
					CPU:      lastMetric.CPUUsagePercent,
					MemUsed:  lastMetric.MemUsedBytes,
					MemLimit: lastMetric.MemLimitBytes,
				})
			}
		}

		database.DB.Raw(`
			SELECT upstream_addr, server_name, status, SUM(requests_count) as requests_count
			FROM metric_load_balancers
			WHERE timestamp >= NOW() - INTERVAL '5 seconds'
			GROUP BY upstream_addr, server_name, status
		`).Scan(&res.LoadBalancing)

		json.NewEncoder(w).Encode(res)
	}))

	log.Printf("🚀 Servidor da API rodando em http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Erro crítico na API HTTP: %v", err)
	}
}
