package main

import (
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"PokeGo/constants"
	"PokeGo/entities"
)

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	spriteRect := image.Rect(
		int(sprite.X),
		int(sprite.Y),
		int(sprite.X)+constants.SpriteSize,
		int(sprite.Y)+constants.SpriteSize,
	)

	for _, collider := range colliders {
		if spriteRect.Overlaps(collider) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - constants.SpriteSize
			} else if sprite.Dx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
			sprite.Dx = 0.0
		}
	}
}

func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	spriteRect := image.Rect(
		int(sprite.X),
		int(sprite.Y),
		int(sprite.X)+constants.SpriteSize,
		int(sprite.Y)+constants.SpriteSize,
	)

	for _, collider := range colliders {
		if spriteRect.Overlaps(collider) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - constants.SpriteSize
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
			sprite.Dy = 0.0
		}
	}
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Pokemon Go")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
