package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/infobsmi/bsmi-go/idle_conn"
	"github.com/panjf2000/ants/v2"
)

var (
	l           string
	r           string
	DialTimeout = 2 * time.Second
	IdleTimeout = 20 * time.Second
	ap          *ants.Pool
)

func handler(conn net.Conn, r string) {
	dialer := &net.Dialer{
		Timeout:   DialTimeout,
		KeepAlive: IdleTimeout,
	}

	client, err := dialer.Dial("tcp", r)
	if err != nil {
		fmt.Println("Dial remote failed", err)

		return
	}
	fmt.Println("To: Connected to remote ", r)

	idleCbw := idle_conn.NewIdleConnWithIdleTimeOut(conn, IdleTimeout)
	idleCbr := idle_conn.NewIdleConnWithIdleTimeOut(client, IdleTimeout)
	doneC := make(chan bool, 2)
	_ = ap.Submit(func() { copySync(idleCbw, idleCbr, doneC) })
	_ = ap.Submit(func() { copySync(idleCbr, idleCbw, doneC) })
	<-doneC
	fmt.Println(" finish copy")
	if doneC != nil {
		close(doneC)
		doneC = nil
	}
	if conn != nil {
		defer conn.Close()
	}
	if client != nil {
		defer client.Close()
	}
	fmt.Println(" close connections")
}

func handlerMultiple(conn net.Conn, r []string) {
	dialer := &net.Dialer{
		Timeout:   DialTimeout,
		KeepAlive: IdleTimeout,
	}

	primaryClient, err := dialer.Dial("tcp", r[0])
	if err != nil {
		fmt.Println("Dial remote failed", err)

		return
	}
	fmt.Println("To: Connected to remote ", r)

	var multiCbw []io.Writer
	for _, k := range r {
		kc, err := dialer.Dial("tcp", k)
		if err != nil {
			fmt.Println("Dial remote failed: ", err)
			continue
		}
		multiCbw = append(multiCbw, kc)
	}
	multiCbw = append(multiCbw, primaryClient)

	idleInput := idle_conn.NewIdleConnWithIdleTimeOut(conn, IdleTimeout)

	idleCbw := io.MultiWriter(multiCbw...)

	idleCbr := idle_conn.NewIdleConnWithIdleTimeOut(primaryClient, IdleTimeout)

	doneC := make(chan bool, 2)
	// 回来，从远程 到 主线程

	// 去程：从主线程 到 多端
	go copySync(idleCbw, idleInput, doneC)
	go copySyncMultiple(idleInput, idleCbr, doneC)
	<-doneC
	<-doneC
	fmt.Println(" finish copy")
	if conn != nil {
		defer conn.Close()
	}
	fmt.Println(" close connections")
}

func copySyncMultiple(w io.Writer, r io.Reader, doneC chan<- bool) {
	if _, err := io.Copy(w, r); err != nil && err != io.EOF {
		fmt.Printf(" failed to copy : %v\n", err)
	}

	fmt.Printf(" finished copying\n")
	doneC <- true

}
func copySync(w io.Writer, r io.Reader, doneC chan<- bool) {
	if _, err := io.Copy(w, r); err != nil && err != io.EOF {
		fmt.Printf(" failed to copy : %v\n", err)
	}

	fmt.Printf(" finished copying\n")
	doneC <- true

}
func main() {
	flag.StringVar(&l, "l", "", "listen host:port")
	flag.StringVar(&r, "r", "", "remote host:port, if mutiple, split with , the first is the primary target")
	flag.Parse()
	if len(l) <= 0 {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	if len(r) <= 0 {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	ap, _ = ants.NewPool(2000, ants.WithNonblocking(true))

	for i := 0; i < 20; i++ {
		tp := i
		_ = ap.Submit(func() {
			fmt.Printf("预热antsPool: %d\n ", tp)
		})
	}
	fmt.Println("Listen on:", l)
	fmt.Println("Forward request to:", r)
	listener, err := net.Listen("tcp", l)

	fmt.Println("Dial timeout: ", DialTimeout)
	if err != nil {
		fmt.Println("Failed to listen on ", l, err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept listener. ", err)
			return
		}
		fmt.Println("From: Accepted connection: ", conn.RemoteAddr().String())
		//go handler(conn, r)
		if strings.Contains(r, ",") {
			_ = ap.Submit(func() { handlerMultiple(conn, strings.Split(r, ",")) })
		} else {

			_ = ap.Submit(func() { handler(conn, r) })
		}
	}
}
