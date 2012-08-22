package main

import (
	"bytes"
	"log"
	"net"
	"strconv"
)

type CollectionServer struct {
	metrics *Metrics
}

func NewCollectionServer(metrics *Metrics) *CollectionServer {
	return &CollectionServer{metrics: metrics}
}

func (server *CollectionServer) ListenAndServe(address string) {
	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		log.Fatalf("CollectionServer.ListenAndServe: %s", err.Error())
	}

	log.Println("server is up")

	for {
		packet := make([]byte, 512)
		n, _, err := conn.ReadFrom(packet)

		if err != nil {
			return
		}

		server.handlePacket(packet[:n])
	}
}

func (server *CollectionServer) handlePacket(packet []byte) {
	server.metrics.UpdateCounter("statsd.packets_received", 1, 1)

	metrics := bytes.Split(packet, []byte("\n"))

	for _, metric := range metrics {
		parts := bytes.Split(metric, []byte(":"))
		key := string(parts[0])

		for _, bit := range parts[1:] {
			fields := bytes.Split(bit, []byte("|"))

			if len(fields) == 1 {
				server.metrics.UpdateCounter("statsd.bad_lines_seen", 1, 1)
				continue
			}

			sampleRate := float64(1)
			value, _ := strconv.ParseFloat(string(fields[0]), 64)

			if len(fields) == 3 {
				sampleRate, _ = strconv.ParseFloat(string(fields[2]), 64)
			}

			switch {
			case string(fields[1]) == "ms":
				server.metrics.UpdateTimer(key, value)
			case string(fields[1]) == "g":
				server.metrics.UpdateGauge(key, value)
			default:
				server.metrics.UpdateCounter(key, value, sampleRate)
			}
		}
	}
}
