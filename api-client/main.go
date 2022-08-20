package main

import (
	"apiclient/pkg/config"
	"apiclient/pkg/slug"
	"apiclient/pkg/udm"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type hostRecord struct {
	macAddress string
	name string
	ipAddress string
}

func main() {
	config := config.InitConfig()
	udmClient := udm.New(udm.UdmConfig{
		Address: config.Address,
		Username: config.Username,
		Password: config.Password,
		Site: config.Site,
	})

	foundClients := udmClient.GetConfiguredClients()
	if !config.FixedOnly {
		foundClients = append(foundClients, udmClient.GetActiveClients()...)
	}

	networkClients := reduceNetworkClients(foundClients, config)

	outputHostAliases(config.HostAliases)
	outputRecords(networkClients)
}

// Reduce clients list into a map indexed by mac address and removing
// any clients with invalid names or ip addresses.
func reduceNetworkClients(clients []udm.NetworkClient, config *config.Configuration) map[string]hostRecord {
	networkClients := make(map[string]hostRecord)
	lowercaseHostnames := config.LowercaseHostnames

	for _, networkClient := range clients {
		name := networkClient.Name
		if name == "" {
			name = networkClient.Hostname
		}
		name = slug.Hostname(name)
		if len(name) < 1 {
			// It is possible that an entry has an empty string set for both
			// "hostname" and "name" in the UDM data. In that case, there's not
			// much we can do. We _could_ make up a name based on the MAC, or other
			// data, but at this time, we just do not care.
			continue
		}
		if lowercaseHostnames {
			name = strings.ToLower(name)
		}

		ip := networkClient.IpAddress
		if len(networkClient.FixedIpAddress) > 0 {
			ip = networkClient.FixedIpAddress
		}
		if len(ip) < 1 {
			continue
		}

		mac := networkClient.MacAddress

		networkClients[mac] = hostRecord{
			macAddress: mac,
			name: name,
			ipAddress: ip,
		}
	}

	return networkClients
}

func outputHostAliases(aliases []config.HostAlias) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprint(tw, "# host aliases\n")
	for _, v := range aliases {
		fmt.Fprintf(tw, "%s\t%s", v.IpAddress, v.Name)
	}
	fmt.Fprint(tw, "\n\n")
	tw.Flush()
}

func outputRecords(records map[string]hostRecord) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprint(tw, "# UDM records\n")
	for _, v := range records {
		fmt.Fprintf(tw, "%s\t%s\t# %s\n", v.ipAddress, v.name, v.macAddress)
	}
	tw.Flush()
}
