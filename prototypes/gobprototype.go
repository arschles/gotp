package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
)

type sendable struct {
	Number  int
	String  string
	Boolean bool
}

func decode(conn net.Conn) {
	decoderPtr := gob.NewDecoder(conn)
	snd := sendable{}
	err := decoderPtr.Decode(&snd)
	if err != nil {
		log.Fatalln("error decoding", err)
	} else {
		log.Println("decoded", snd)
	}
}

func main() {
	client := flag.Bool("client", false, "true if this test should be the client. defaults to the server")
	port := flag.Int("port", 8080, "the port that this test should use")
	flag.Parse()
	if *client {
		log.Println("running client on port", *port)
	} else {
		log.Println("running server on port", *port)
	}

	netString := fmt.Sprintf("localhost:%v", *port)

	if *client {
		log.Println("sending data on port", port)
		conn, err := net.Dial("tcp", netString)
		if err != nil {
			log.Fatalln("error dialing", err)
			panic(err)
		}

		encoderPtr := gob.NewEncoder(conn)
		snd := sendable{Number: 2, String: "abc", Boolean: true}
		err = encoderPtr.Encode(snd)
		if err != nil {
			log.Fatalln("error encoding", err)
			panic(err)
		}
		log.Println("sent", snd)
	} else {
		log.Println("receiving data on port", *port)
		ln, err := net.Listen("tcp", netString)
		if err != nil {
			log.Fatalln("error listening", err)
			panic(err)
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalln("error accepting", err)
			}
			go decode(conn)

		}
	}
}
