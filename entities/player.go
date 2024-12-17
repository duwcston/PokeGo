package entities

import (
	"PokeGo/animations"

	"github.com/hajimehoshi/ebiten/v2"
)

type PlayerState uint8

const (
	Down PlayerState = iota
	Left
	Right
	Up
)

type Player struct {
	ID         string
	Img        *ebiten.Image
	X          float64
	Y          float64
	Dx         float64
	Dy         float64
	Animations map[PlayerState]*animations.Animation
}

func (p *Player) ActiveAnimation(dx, dy int) *animations.Animation {
	if dx > 0 {
		return p.Animations[Right]
	}
	if dx < 0 {
		return p.Animations[Left]
	}
	if dy > 0 {
		return p.Animations[Down]
	}
	if dy < 0 {
		return p.Animations[Up]
	}
	return nil
}
