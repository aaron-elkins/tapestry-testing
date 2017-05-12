package main

import (
	"fmt"
	"net"
	"pf"
	"strconv"
)

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
		go handleConn(conn, srcIP, srcPort, rIP, rPort)
	}
}
