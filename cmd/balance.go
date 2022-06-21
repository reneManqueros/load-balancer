package cmd

import (
	"github.com/spf13/cobra"
	"loadbalancer/models"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Run load balancer",
	Run: func(cmd *cobra.Command, args []string) {

		rand.Seed(time.Now().UnixMilli())

		// Defaults
		listenNetwork := "tcp"
		listenAddress := ":8081"
		managementAddress := ":33333"
		configFile := "./backends.yml"
		isVerbose := false
		timeout := 0
		connections := 0
		// Overrides
		for _, v := range args {
			argumentParts := strings.Split(v, "=")
			if len(argumentParts) == 2 {
				if argumentParts[0] == "network" {
					listenNetwork = argumentParts[1]
				}
				if argumentParts[0] == "address" {
					listenAddress = argumentParts[1]
				}
				if argumentParts[0] == "management" {
					managementAddress = argumentParts[1]
				}
				if argumentParts[0] == "config" {
					configFile = argumentParts[1]
				}
				if argumentParts[0] == "timeout" {
					timeout, _ = strconv.Atoi(argumentParts[1])
				}
				if argumentParts[0] == "connections" {
					connections, _ = strconv.Atoi(argumentParts[1])
				}
				if argumentParts[0] == "verbose" && argumentParts[1] == "true" {
					isVerbose = true
				}
			}
		}

		lb := models.LoadBalancer{
			Network:     listenNetwork,
			Timeout:     timeout,
			Source:      listenAddress,
			Mutex:       sync.Mutex{},
			ConfigFile:  configFile,
			IsVerbose:   isVerbose,
			Connections: connections,
		}
		lb.FromDisk()

		if managementAddress != "" {
			go models.Management{
				ListenAddress: managementAddress,
				LoadBalancer:  &lb,
			}.Listen()
		}

		lb.Listen()
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags()
}
