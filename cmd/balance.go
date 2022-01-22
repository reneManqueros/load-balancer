package cmd

import (
	"github.com/spf13/cobra"
	"loadbalancer/models"
	"math/rand"
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
			}
		}

		lb := models.LoadBalancer{
			Network:    listenNetwork,
			Source:     listenAddress,
			Mutex:      sync.Mutex{},
			ConfigFile: configFile,
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
