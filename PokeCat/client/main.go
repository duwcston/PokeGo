package main

import (
	"bufio"
	"fmt"
	"image"
	"log"
	"net"
	"os"

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

var conn net.Conn

func main() {
	var err error
	conn, err = net.Dial("tcp", "localhost:3000")
	if err != nil {
		fmt.Println("Connection error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println("Server:", scanner.Text())
		}
	}()

	game := NewGame()

	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetWindowTitle("Pokémon Multiplayer Game")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game.spawnPokeballs()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
