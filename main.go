package main

import "flag"

var (
	collectionAddress = flag.String("address", ":8125", "address for UDP messages")
	debug             = flag.Bool("debug", false, "turn on debugging")
	flushInterval     = flag.Int("flush-interval", 10, "interval to flush statistics")
	graphiteAddress   = flag.String("graphite", "", "graphite service address")
	managementAddress = flag.String("management", ":8126", "address for the TCP management interface")
)

func main() {
	flag.Parse()

	metrics := NewMetrics()

	managementServer := NewManagementServer(metrics)
	go managementServer.ListenAndServe(*managementAddress)

	collectionServer := NewCollectionServer(metrics)
	collectionServer.ListenAndServe(*collectionAddress)
}
