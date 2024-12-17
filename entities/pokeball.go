package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Pokeball struct {
	Img *ebiten.Image
	X   float64
	Y   float64
}
