package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	Img *ebiten.Image
	X   float64
	Y   float64
	Dx  float64
	Dy  float64
}
