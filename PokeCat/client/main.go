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

func CheckCollisionHorizontal(player *entities.Player, colliders []image.Rectangle) {
	spriteRect := image.Rect(
		int(player.X),
		int(player.Y),
		int(player.X)+constants.SpriteSize,
		int(player.Y)+constants.SpriteSize,
	)

	for _, collider := range colliders {
		if spriteRect.Overlaps(collider) {
			if player.Dx > 0.0 {
				player.X = float64(collider.Min.X) - constants.SpriteSize
			} else if player.Dx < 0.0 {
				player.X = float64(collider.Max.X)
			}
			player.Dx = 0.0
		}
	}
}

func CheckCollisionVertical(player *entities.Player, colliders []image.Rectangle) {
	spriteRect := image.Rect(
		int(player.X),
		int(player.Y),
		int(player.X)+constants.SpriteSize,
		int(player.Y)+constants.SpriteSize,
	)

	for _, collider := range colliders {
		if spriteRect.Overlaps(collider) {
			if player.Dy > 0.0 {
				player.Y = float64(collider.Min.Y) - constants.SpriteSize
			} else if player.Dy < 0.0 {
				player.Y = float64(collider.Max.Y)
			}
			player.Dy = 0.0
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

	fmt.Println("Connected to server")

	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := scanner.Text()
	conn.Write([]byte("JOIN AS " + name + "\n"))

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
