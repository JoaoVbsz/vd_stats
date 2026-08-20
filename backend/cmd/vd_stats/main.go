package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/joaov/vd_stats/internal/api"
	"github.com/joaov/vd_stats/internal/database"
	"github.com/joaov/vd_stats/internal/ssh"
)

func main() {
	log.Println("Iniciando motor DockKeeper...")

	_ = godotenv.Load("../.env", ".env")

	err := database.Connect()
	if err != nil {
		log.Fatalf("Falha crítica ao conectar no banco: %v", err)
	}

	go api.StartServer(":8080")

	sshUser := os.Getenv("SSH_USER")
	if sshUser == "" {
		sshUser = "root"
	}
	sshKey := os.Getenv("SSH_KEY_PATH")

	// Seed from .env (Legacy / First run fallback)
	targetVpsList := os.Getenv("TARGET_VPS_IPS")
	if targetVpsList != "" {
		vpsIPs := strings.Split(targetVpsList, ",")
		for _, ip := range vpsIPs {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				name := "VPS Node"
				var server database.Server
				database.DB.Where("host_ip = ?", ip).FirstOrCreate(&server, database.Server{
					Name: name, HostIP: ip, User: sshUser,
				})
			}
		}
	}

	// Load Balancer (Legacy / First run fallback)
	lbIP := os.Getenv("LB_IP")
	if lbIP != "" {
		var server database.Server
		database.DB.Where("host_ip = ?", lbIP).FirstOrCreate(&server, database.Server{
			Name: "Load Balancer", HostIP: lbIP, User: sshUser,
		})
	}

	// Iniciar Streams SSH de todos os servidores cadastrados no Banco
	var servers []database.Server
	database.DB.Find(&servers)

	log.Printf("Carregados %d servidores do banco de dados.", len(servers))
	for _, s := range servers {
		if s.Name == "Load Balancer" {
			go func(host, user, key string) {
				for {
					err := ssh.StartNginxStream(host, user, key)
					if err != nil {
						log.Printf("Erro no NGINX Stream %s: %v. Tentando reconectar...", host, err)
						time.Sleep(5 * time.Second)
					}
				}
			}(s.HostIP, s.User, sshKey)
		} else {
			ssh.Manager.Start(s.ID, s.HostIP, s.User, sshKey)
		}
	}

	select {}
}
