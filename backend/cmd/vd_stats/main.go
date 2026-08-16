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
	log.Println("Iniciando motor vd_stats...")

	_ = godotenv.Load("../.env", ".env")

	err := database.Connect()
	if err != nil {
		log.Fatalf("Falha crítica ao conectar no banco: %v", err)
	}

	go api.StartServer(":8080")

	targetVpsList := os.Getenv("TARGET_VPS_IPS")
	lbIP := os.Getenv("LB_IP")
	sshUser := os.Getenv("SSH_USER")
	sshKey := os.Getenv("SSH_KEY_PATH")

	vpsIPs := strings.Split(targetVpsList, ",")

	for _, ip := range vpsIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		go func(h, u, k string) {
			for {
				err := ssh.StartStream(h, u, k)
				if err != nil {
					log.Printf("Erro na VPS %s: %v. Tentando reconectar em 5s...", h, err)
					time.Sleep(5 * time.Second)
				}
			}
		}(ip, sshUser, sshKey)
	}

	if lbIP != "" {
		go func() {
			for {
				err := ssh.StartNginxStream(lbIP, sshUser, sshKey)
				if err != nil {
					log.Printf("Erro no Load Balancer: %v. Tentando reconectar em 5s...", err)
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}

	select {}
}
