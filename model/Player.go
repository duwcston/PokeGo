package model

import "net"

type Player struct {
	Name      string       `json:"Name"`
	Addr      *net.UDPAddr `json:"Addr"`
	Pokeballs int          `json:"Pokeballs"`
	Inventory []Pokemon    `json:"Inventory"`
}
