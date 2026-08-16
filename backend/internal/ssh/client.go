package ssh

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joaov/vd_stats/internal/database"
	"golang.org/x/crypto/ssh"
)

func parseDockerSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" { return 0 }
	var val float64
	var unit string
	fmt.Sscanf(sizeStr, "%f%s", &val, &unit)

	unit = strings.ToUpper(unit)
	multiplier := float64(1)
	switch {
	case strings.Contains(unit, "K"): multiplier = 1024
	case strings.Contains(unit, "M"): multiplier = 1024 * 1024
	case strings.Contains(unit, "G"): multiplier = 1024 * 1024 * 1024
	case strings.Contains(unit, "T"): multiplier = 1024 * 1024 * 1024 * 1024
	}
	return int64(val * multiplier)
}

func parsePercent(p string) float64 {
	p = strings.TrimSuffix(strings.TrimSpace(p), "%")
	val, _ := strconv.ParseFloat(p, 64)
	return val
}

type SysPayload struct {
	Uptime     float64               `json:"uptime"`
	DiskRoot   string                `json:"disk_root"`
	Containers []map[string]string   `json:"containers"`
}

func StartStream(host, user, keyPath string) error {
	name := "VPS Veloci-Auto"
	if strings.Contains(host, "39") {
		name = "VPS Veloci-BI"
	}
	var server database.Server
	database.DB.Where("host_ip = ?", host).FirstOrCreate(&server, database.Server{
		Name: name, HostIP: host, User: user,
	})

	keyPath = strings.Replace(keyPath, "~", os.Getenv("HOME"), 1)
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil { return err }

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil { return err }

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 10 * time.Second,
	}

	hostPort := fmt.Sprintf("%s:22", host)
	log.Printf("🔌 Iniciando conexão SSH com a VPS %s...", hostPort)
	
	startPing := time.Now()
	client, err := ssh.Dial("tcp", hostPort, config)
	if err != nil { return err }
	defer client.Close()
	
	latencyMs := float64(time.Since(startPing).Milliseconds())

	session, err := client.NewSession()
	if err != nil { return err }
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil { return err }

	query := `while true; do
  DOCKER_JSON=$(docker stats --no-stream --format '{"docker_id":"{{.ID}}","name":"{{.Name}}","cpu_percent":"{{.CPUPerc}}","mem_usage":"{{.MemUsage}}"}' | paste -sd, -)
  UPTIME=$(cat /proc/uptime | awk '{print $1}')
  DISK_ROOT=$(df -B1 / | awk 'NR==2 {print $3","$2}')
  echo "{\"uptime\":$UPTIME,\"disk_root\":\"$DISK_ROOT\",\"containers\":[$DOCKER_JSON]}"
  sleep 2
done`

	err = session.Start(query)
	if err != nil { return err }

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		
		var payload SysPayload
		if err := json.Unmarshal([]byte(line), &payload); err != nil { continue }

		diskParts := strings.Split(payload.DiskRoot, ",")
		var dUsed, dTotal int64
		if len(diskParts) == 2 {
			dUsed, _ = strconv.ParseInt(diskParts[0], 10, 64)
			dTotal, _ = strconv.ParseInt(diskParts[1], 10, 64)
		}

		database.DB.Create(&database.MetricServer{
			ServerID: server.ID, UptimeSeconds: payload.Uptime,
			DiskUsedBytes: dUsed, DiskTotalBytes: dTotal,
			PingLatencyMs: latencyMs, Timestamp: time.Now().UTC(),
		})

		for _, raw := range payload.Containers {
			dockerID := raw["docker_id"]
			var container database.Container
			database.DB.Where("server_id = ? AND docker_id = ?", server.ID, dockerID).FirstOrCreate(&container, database.Container{
				ServerID: server.ID, DockerID: dockerID, Name: raw["name"],
			})

			memParts := strings.Split(raw["mem_usage"], "/")
			var memUsed, memLimit int64
			if len(memParts) == 2 {
				memUsed = parseDockerSize(memParts[0])
				memLimit = parseDockerSize(memParts[1])
			}

			database.DB.Create(&database.MetricContainer{
				ContainerID: container.ID, CPUUsagePercent: parsePercent(raw["cpu_percent"]),
				MemUsedBytes: memUsed, MemLimitBytes: memLimit,
				Status: "running", Timestamp: time.Now().UTC(),
			})
		}
		log.Printf("[GRAVADO] %s | Latência: %.0fms | Containers: %d", host, latencyMs, len(payload.Containers))
	}
	return session.Wait()
}

func StartNginxStream(host, user, keyPath string) error {
	name := "Load Balancer"
	var server database.Server
	database.DB.Where("host_ip = ?", host).FirstOrCreate(&server, database.Server{
		Name: name, HostIP: host, User: user,
	})

	keyPath = strings.Replace(keyPath, "~", os.Getenv("HOME"), 1)
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil { return err }

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil { return err }

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 10 * time.Second,
	}

	hostPort := fmt.Sprintf("%s:22", host)
	log.Printf("🔌 Iniciando stream do NGINX na VPS %s...", hostPort)
	
	client, err := ssh.Dial("tcp", hostPort, config)
	if err != nil { return err }
	defer client.Close()

	session, err := client.NewSession()
	if err != nil { return err }
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil { return err }

	query := `tail -n 0 -F /var/log/nginx/access.log`
	err = session.Start(query)
	if err != nil { return err }

	scanner := bufio.NewScanner(stdout)
	
	type LbKey struct {
		Upstream   string
		ServerName string
		Status     string
	}
	var mu sync.Mutex
	counts := make(map[LbKey]int)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			if len(counts) > 0 {
				for key, count := range counts {
					database.DB.Create(&database.MetricLoadBalancer{
						UpstreamAddr:  key.Upstream,
						ServerName:    key.ServerName,
						Status:        key.Status,
						RequestsCount: count,
						Timestamp:     time.Now().UTC(),
					})
				}
				counts = make(map[LbKey]int)
			}
			mu.Unlock()
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		idxTo := strings.Index(line, " to: ")
		if idxTo != -1 {
			prefix := line[:idxTo]
			partsPrefix := strings.Split(prefix, " - ")
			serverName := ""
			if len(partsPrefix) > 0 {
				serverName = strings.TrimSpace(partsPrefix[len(partsPrefix)-1])
			}

			rem := line[idxTo+5:]
			parts := strings.SplitN(rem, ": ", 2)
			if len(parts) >= 2 {
				upstream := strings.TrimSpace(parts[0])
				if upstream == "-" {
					upstream = "Local (Nginx/Cache)"
				}
				if upstream != "" {
					status := "200"
					reqStatusStr := parts[1]
					if strings.Contains(reqStatusStr, " 50") || strings.Contains(reqStatusStr, " 40") {
						if strings.Contains(reqStatusStr, " 500") { status = "500" }
						if strings.Contains(reqStatusStr, " 502") { status = "502" }
						if strings.Contains(reqStatusStr, " 503") { status = "503" }
						if strings.Contains(reqStatusStr, " 504") { status = "504" }
						if strings.Contains(reqStatusStr, " 404") { status = "404" }
						if strings.Contains(reqStatusStr, " 400") { status = "400" }
						if strings.Contains(reqStatusStr, " 429") { status = "429" }
					}

					key := LbKey{Upstream: upstream, ServerName: serverName, Status: status}
					mu.Lock()
					counts[key]++
					mu.Unlock()
				}
			}
		}
	}
	return session.Wait()
}
