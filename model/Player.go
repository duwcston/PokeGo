package model

import "net"

type Player struct {
	Name      string
	Addr      *net.UDPAddr
	Inventory []Pokemon
}
