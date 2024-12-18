package model

import "net"

type Player struct {
	Name      string       `json:"Name"`
	Addr      *net.UDPAddr `json:"Addr"`
	Inventory []Pokemon    `json:"Inventory"`
}
