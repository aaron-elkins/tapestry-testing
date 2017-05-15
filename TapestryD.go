package main

import (
	"fmt"
	psunet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"net"
	"os/exec"
	"pf"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// embed regexp.Regexp in a new type so we can extend it
type myRegexp struct {
	*regexp.Regexp
}

// add a new method to our new regular expression type
func (r *myRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}

func getNatTable() string {
	out, err := exec.Command("pfctl", "-ss").Output()
	if err != nil {
		fmt.Printf("Read NAT table failed: %s", err.Error())
	}
	return string(out)
}

// Return empty string if not found
func getOriginDest(king string, natTable string, src string, dst string) string {
	lines := strings.Split(natTable, "\n")
	p := king + ` (?P<dst>.+) <- (?P<orig>.+) <- (?P<src>\S+)`
	re := myRegexp{regexp.MustCompile(p)}

	for _, line := range lines {
		m := re.FindStringSubmatchMap(line)
		if m != nil {
			s := m["src"]
			d := m["dst"]
			orig := m["orig"]
			if s == src && d == dst {
				return orig
			}
		}
	}

	return ""
}

func reversePids(pids []int32) {
	for i, j := 0, len(pids)-1; i < j; i, j = i+1, j-1 {
		pids[i], pids[j] = pids[j], pids[i]
	}
}

func getProcessIDByPort(proto string, ip string, port int) int32 {
	var rPid int32 = 0
	pids, err := process.Pids()

	if err != nil {
		fmt.Printf("Can not get PIDs, %s\n", err.Error())
		return 0
	}

	reversePids(pids)

	for _, pid := range pids {
		connectionStats, err := psunet.ConnectionsPid(proto, pid)
		if err != nil {
			fmt.Printf("Can not get connections for PID:%d, %s\n", pid, err.Error())
			continue
		}
		for _, stat := range connectionStats {
			if ip == stat.Laddr.IP && uint32(port) == stat.Laddr.Port {
				rPid = pid
				return rPid
			}
		}

	}
	return 0
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

func readUDP(udpServer *net.UDPConn) {
	buf := make([]byte, 1024)

	for {
		n, addr, err := udpServer.ReadFromUDP(buf)

		srcAddr := addr
		dstAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:11235")

		fmt.Printf("Src: %s Dst: %s\n", srcAddr.String(), dstAddr.String())

		natTable := getNatTable()
		orig := getOriginDest("udp", natTable, srcAddr.String(), dstAddr.String())

		fmt.Printf("Remote IP: %s\n", orig)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func main() {
	fmt.Println("Starting server...")

	fmt.Printf("TCP: %d UDP: %d\n", pf.IPPROTO_TCP, pf.IPPROTO_UDP)
	// Listen TCP on 0.0.0.0:11235
	ln, _ := net.Listen("tcp", "0.0.0.0:11235")

	// Listen UDP on 0.0.0.0:11235
	localUDPAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:11235")
	udpServer, _ := net.ListenUDP("udp4", localUDPAddr)

	go readUDP(udpServer)

	for {

		conn, _ := ln.Accept()

		srcAddr := conn.RemoteAddr()
		destAddr := conn.LocalAddr()

		fmt.Printf("[TCP] Src: %s Dst: %s\n", srcAddr, destAddr)
		srcIP := srcAddr.(*net.TCPAddr).IP
		srcPort := srcAddr.(*net.TCPAddr).Port

		destIP := destAddr.(*net.TCPAddr).IP
		destPort := destAddr.(*net.TCPAddr).Port
		rIP, rPort, err := pf.QueryNat(pf.AF_INET, pf.IPPROTO_TCP, srcIP, srcPort, destIP, destPort)

		if err != nil {
			fmt.Printf("Query Nat fail! (TCP) %s\n", err.Error())
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
