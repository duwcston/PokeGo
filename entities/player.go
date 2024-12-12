package entities

import "PokeGo/animations"

type PlayerState uint8

const (
	Down PlayerState = iota
	Left
	Right
	Up
)

type Player struct {
	*Sprite
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
