package main

import (
	"fmt"
	psunet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"net"
	"pf"
	"strconv"
	"time"
)

func reversePids(pids []int32) {
	for i, j := 0, len(pids)-1; i < j; i, j = i+1, j-1 {
		pids[i], pids[j] = pids[j], pids[i]
	}
}

func getProcessIDByPort(proto string, ip string, port int) int32 {
	pids, _ := process.Pids()

	reversePids(pids)

	var rPid int32 = 0
	for _, pid := range pids {
		fmt.Printf("PID: %d\n", pid)
		connectionStats, _ := psunet.ConnectionsPid(proto, pid)
		for _, stat := range connectionStats {
			if ip == stat.Laddr.IP && uint32(port) == stat.Laddr.Port {
				rPid = pid
				return rPid
			}
		}

	}
	return rPid
}

func connToConn(src net.Conn, dst net.Conn) {
	for {
		data := make([]byte, 1024*10)
		read, err := src.Read(data)
		if err != nil {
			fmt.Println("Conn Closed\n")
			break
		}

		dst.Write(data[:read])
	}
}

func handleConn(conn net.Conn, src net.IP, srcPort int, dst net.IP, dstPort int) {
	addr := dst.String() + ":" + strconv.Itoa(dstPort)

	dstConn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Connect to remote failed: %s\n", addr)
		return
	}
	go connToConn(conn, dstConn)
	go connToConn(dstConn, conn)
}

func main() {
	fmt.Println("Starting server...")

	// Listen on 127.0.0.1:11235
	ln, _ := net.Listen("tcp", "0.0.0.0:11235")

	for {

		conn, _ := ln.Accept()

		srcAddr := conn.RemoteAddr()
		destAddr := conn.LocalAddr()

		srcIP := srcAddr.(*net.TCPAddr).IP
		srcPort := srcAddr.(*net.TCPAddr).Port

		destIP := destAddr.(*net.TCPAddr).IP
		destPort := destAddr.(*net.TCPAddr).Port

		rIP, rPort, err := pf.QueryNat(pf.AF_INET, pf.IPPROTO_TCP, srcIP, srcPort, destIP, destPort)

		if err != nil {
			fmt.Println("Query Nat fail!")
			continue
		}

		fmt.Println("Handle connection:" + conn.RemoteAddr().String() + "=>" + rIP.String() + ":" + strconv.Itoa(rPort))
		start := time.Now()
		pid := getProcessIDByPort("tcp", srcIP.String(), srcPort)
		elapsed := time.Since(start)
		fmt.Printf("getProcessIDByPort() took %s\n", elapsed)
		fmt.Printf("PID: %d want to connect to %s:%d\n", pid, rIP, rPort)
		go handleConn(conn, srcIP, srcPort, rIP, rPort)
	}
}
