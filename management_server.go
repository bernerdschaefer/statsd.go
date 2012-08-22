package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type ManagementServer struct {
	sock    net.Listener
	metrics *Metrics
}

func NewManagementServer(metrics *Metrics) *ManagementServer {
	return &ManagementServer{metrics: metrics}
}

func (server *ManagementServer) ListenAndServe(address string) {
	sock, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatalf("ManagementServer.ListenAndServe: %s", err.Error())
	}

	server.sock = sock

	for {
		conn, err := sock.Accept()

		if err != nil {
			continue
		}

		go server.handleConnection(conn)
	}
}

type command [][]byte

func (c command) Verb() (verb string) {
	if len(c) > 0 {
		verb = string(c[0])
	}
	return verb
}

func (c command) Args() (args []string) {
	if len(c) > 1 {
		for _, arg := range c[1:] {
			args = append(args, string(arg))
		}
	}
	return args
}

func (server *ManagementServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := bufio.NewReader(conn)
	out := bufio.NewWriter(conn)
	serializer := json.NewEncoder(out)

	for {
		line, _, err := buf.ReadLine()

		if err != nil {
			return
		}

		cmd := command(bytes.Fields(line))

		switch cmd.Verb() {
		case "help":
			out.WriteString("Commands: stats, counters, timers, gauges, delcounters, deltimers, delgauges, quit\n\n")
		case "stats":
			counters, gauges, timers := server.metrics.Read()
			for key, val := range counters {
				fmt.Fprintf(out, "%s.%s: %v\n", "counters", key, val)
			}
			for key, val := range gauges {
				fmt.Fprintf(out, "%s.%s: %v\n", "gauges", key, val)
			}
			for key, val := range timers {
				fmt.Fprintf(out, "%s.%s: %v\n", "timers", key, val)
			}
			out.WriteString("END\n\n")
		case "counters":
			counters, _, _ := server.metrics.Read()
			serializer.Encode(counters)
			out.WriteString("END\n\n")
		case "gauges":
			_, gauges, _ := server.metrics.Read()
			serializer.Encode(gauges)
			out.WriteString("END\n\n")
		case "timers":
			_, _, timers := server.metrics.Read()
			serializer.Encode(timers)
			out.WriteString("END\n\n")
		case "delcounters":
			for _, key := range cmd.Args() {
				server.metrics.DeleteCounter(key)
				out.WriteString("deleted: " + key + "\n")
			}
			out.WriteString("END\n\n")
		case "delgauges":
			for _, key := range cmd.Args() {
				server.metrics.DeleteGauge(key)
				out.WriteString("deleted: " + key + "\n")
			}
			out.WriteString("END\n\n")
		case "deltimers":
			for _, key := range cmd.Args() {
				server.metrics.DeleteTimer(key)
				out.WriteString("deleted: " + key + "\n")
			}
			out.WriteString("END\n\n")
		case "quit":
			return
		default:
			out.WriteString("ERROR\n")
		}
		out.Flush()
	}
}
