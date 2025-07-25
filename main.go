package main

import (
	"crypto/tls"
	"io"
	"log"
	"os"
)

func main() {
	ssh_gateway := os.Args[1]
	target_server := os.Args[2]

	client, err := tls.Dial("tcp", ssh_gateway, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         target_server,
	})

	if err != nil {
		panic(err)
	}
	defer client.Close()

	transport := sshTransport{
		TlsChannel: client,
		ErrC:       make(chan error, 1),
	}

	go transport.copyFromChannel()
	go transport.copyToChannel()

	err = <-transport.ErrC

	if err != nil {
		os.Stdout.WriteString("Error during copy: " + err.Error() + "\n")
		log.Fatalf("Stream closed: %v", err)
		os.Exit(1)
	}

	os.Stdout.WriteString("Connection closed by server\n")
	os.Exit(0)
}

type sshTransport struct {
	TlsChannel io.ReadWriter
	ErrC       chan error
}

func (c sshTransport) copyFromChannel() {
	_, err := io.Copy(os.Stdout, c.TlsChannel)
	c.ErrC <- err
}

func (c sshTransport) copyToChannel() {
	_, err := io.Copy(c.TlsChannel, os.Stdin)
	c.ErrC <- err
}
