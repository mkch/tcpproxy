package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"strconv"
)

func main() {
	var server, addr string
	var serverPort, port int
	flag.StringVar(&server, "server", "", "TCP server address")
	flag.IntVar(&serverPort, "server-port", 0, "TCP server port")
	flag.StringVar(&addr, "addr", "localhost", "Local address to listen")
	flag.IntVar(&port, "port", 8888, "Local port to listen")
	flag.Parse()

	if server == "" {
		log.Fatal("invalid server")
	}
	if serverPort == 0 {
		log.Fatal("invalid server port")
	}
	if addr == "" {
		log.Fatal("invalid addr")
	}
	if port == 0 {
		log.Fatal("invalid port")
	}

	run(server, serverPort, addr, port)
}

func run(server string, serverPort int, addr string, port int) {
	l, err := net.Listen("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		}
		go serve(conn, server, serverPort)
	}

}

func serve(clientConn net.Conn, server string, serverPort int) {
	defer clientConn.Close()
	// Connect to the server.
	serverConn, err := net.Dial("tcp", net.JoinHostPort(server, strconv.Itoa(serverPort)))
	if err != nil {
		log.Println(err)
		return
	}
	defer serverConn.Close()

	// Read from the server.
	var serverBuf = make(chan []byte)
	go read(serverConn, serverBuf)

	// Read from the client.
	var clientBuf = make(chan []byte)
	go read(clientConn, clientBuf)

	// Forward the data back and forth.
	var buf []byte
	var ok = true
	for {
		var dest net.Conn
		select {
		case buf, ok = <-serverBuf:
			dest = clientConn
		case buf, ok = <-clientBuf:
			dest = serverConn
		}
		if !ok { // Either side closed the connection.
			break
		}

		// Forward the data.
		_, err := dest.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}

// read reads from conn and send the read data to ch,
// until EOF or some error occurred.
// ch is closed when read returns.
func read(conn net.Conn, ch chan<- []byte) {
	defer close(ch)
	for {
		var buf = make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				log.Println(err)
			}
			return
		}
		ch <- buf[:n]
	}
}
