package main

import (
	"fmt"
	"net"
	"time"
	"log"
)

func main() {
	address := "103.156.128.114:9967"
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Fatal("Dial Error:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to", address)

	// Send /login command (encoded in shorter form for test)
	// Word length 6, content "/login", then empty word to terminate sentence? 
	// No, API protocol is: [Len][Content].
	// Len < 0x80 is 1 byte.
	// "/login" is 6 bytes.
	// Termination is empty word (len 0).
	
	// Byte stream: 0x06 (len) "/login" 0x00 (len 0, end of sentence)
	// Actually, just sending /login word might trigger a response (even without empty word termination, usually).
	// But let's follow spec: Sentence = Word + ZeroWord.
	
	payload := []byte{0x06, '/', 'l', 'o', 'g', 'i', 'n', 0x00}
	
	_, err = conn.Write(payload)
	if err != nil {
		log.Fatal("Write Error:", err)
	}

	// Read response
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal("Read Error:", err)
	}

	fmt.Printf("Received %d bytes: %x\n", n, buf[:n])
	fmt.Printf("String: %s\n", string(buf[:n]))
}
