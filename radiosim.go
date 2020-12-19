package main

// radiosim is a program to look like a openHPSDR radio and tangerineSDR radio for the purpose of
// testing the discovery tool
//
// by Dave Larsen KV0S
// code at github.com/kv0s/radiosim
//
// licensed under the GPL3

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Radio    string
	Version  string
	Protocol string
	Status   string
}

type packetBt struct {
	Status   byte
	radioMAC []byte
	Version  byte
	Board    byte
}

func main() {
	var cfg Config
	var radioname string
	//var radioMAC []byte
	var pBt packetBt

	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = toml.Decode(string(data), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Config:\n Radio %s\n Version %#v\n Protocol %s\n status %s\n", cfg.Radio, cfg.Version, cfg.Protocol, cfg.Status)

	if cfg.Status == "idle" {
		pBt.Status = 0x02
	} else {
		pBt.Status = 0x03
	}

	radioname = strings.ToLower(cfg.Radio)
	if radioname == "metis" {
		pBt.Board = 0x00
	} else if radioname == "hermes" {
		pBt.Board = 0x01
	} else if radioname == "griffin" {
		pBt.Board = 0x02
	} else if radioname == "angelia" {
		pBt.Board = 0x04
	} else if radioname == "orian" {
		pBt.Board = 0x05
	} else if radioname == "hermes_lite" {
		pBt.Board = 0x06
	} else if radioname == "tangerinesdr" {
		pBt.Board = 0x07
	}

	// Conver the cfg string to decimal number
	// convert decimal to hex number

	pBt.Version = 0x26

	inf, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", inf)

	for _, i := range inf {
		//Check interface name
		if i.Name == "enp0s25" {
			pBt.radioMAC = i.HardwareAddr
		}
	}

	fmt.Printf(" radioMAC %#v\n", pBt.radioMAC)

	pc, err := net.ListenPacket("udp4", ":1024")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	for {
		buf := make([]byte, 64)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		go handleConnection(pc, n, addr, buf, pBt)
	}
}

// handleConnection is a goroutine to handle on connection at a time
func handleConnection(pc net.PacketConn, n int, addr net.Addr, buf []byte, pkt packetBt) {

	fmt.Printf("Received from %s %x length=%d\n", addr, buf[:n], len(buf))

	rbuf := make([]byte, 64)

	rbuf, err := hex.DecodeString("effe")
	if err != nil {
		panic(err)
	}

	rbuf = append(rbuf, pkt.Status)
	rbuf = append(rbuf, pkt.radioMAC...)
	rbuf = append(rbuf, pkt.Version)
	rbuf = append(rbuf, pkt.Board)

	for i := 1; i < 50; i++ {
		rbuf = append(rbuf, 0x00)
	}

	fmt.Printf("Sent to %s %x length=%d\n", addr, rbuf[:n], len(rbuf))
	pc.WriteTo(rbuf, addr)
}
