package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"
)

func DialSSL(addr string) {

	buf := fmt.Sprintf("Connecting to: %s\r\n", addr)
	AddLine(buf)

	//Todo, allow timeout adjustment and connection canceling.
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		log.Println(err)

		buf := fmt.Sprintf("%s\r\n", err)
		AddLine(buf)
		return
	}
	MainWin.sslCon = conn
}

func readNet() {
	go func() {
		for {
			buf := make([]byte, MAX_INPUT_LENGTH)
			if MainWin.sslCon != nil {
				n, err := MainWin.sslCon.Read(buf)
				if err != nil {
					log.Println(n, err)
					MainWin.sslCon.Close()

					buf := fmt.Sprintf("Lost connection to %s: %s\r\n", MainWin.serverAddr, err)
					AddLine(buf)
					MainWin.sslCon = nil
				}
				newData := string(buf[:n])
				AddLine(newData)
			}
			time.Sleep(time.Millisecond * NET_POLL_MS)
		}
	}()
}
