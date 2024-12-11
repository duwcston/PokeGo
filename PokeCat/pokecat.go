package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"PokeGo/entities"
)

const (
	screenWidth  = 320
	screenHeight = 240
)

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+16.0,
				int(sprite.Y)+16.0,
			),
		) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - 16.0
			} else if sprite.Dx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
		}
	}
}

func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+16.0,
				int(sprite.Y)+16.0,
			),
		) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - 16.0
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.X)
			}
		}
	}
}

type Game struct {
	player      *entities.Player
	pokeball    []*entities.Pokeball
	tilemapJSON *TilemapJSON
	tilemapImg  *ebiten.Image
	camera      *Camera
	colliders   []image.Rectangle
}

func (g *Game) Update() error {
	g.player.Dx = 0
	g.player.Dy = 0

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dy = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dy = 2
	}

	g.player.X += g.player.Dx
	g.player.Y += g.player.Dy

	CheckCollisionHorizontal(g.player.Sprite, g.colliders)
	CheckCollisionVertical(g.player.Sprite, g.colliders)

	for _, pokeball := range g.pokeball {
		if g.player.X == pokeball.X && g.player.Y == pokeball.Y {
			fmt.Println("Pokeball collected!")
		}
	}

	g.camera.FollowPlayer(g.player.X+8, g.player.Y+8, screenWidth, screenHeight)
	g.camera.Constrain(
		float64(g.tilemapJSON.Layers[0].Width)*16,
		float64(g.tilemapJSON.Layers[0].Height)*16,
		screenWidth, screenHeight,
	)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	for _, layer := range g.tilemapJSON.Layers {
		for index, tileID := range layer.Data {
			if tileID == 0 {
				continue
			}

			tileX := float64((index % layer.Width) * 16)
			tileY := float64((index / layer.Width) * 16)

			opts.GeoM.Translate(tileX, tileY)
			opts.GeoM.Translate(g.camera.X, g.camera.Y)

			srcX := float64((tileID-1)%64) * 16
			srcY := float64((tileID-1)/64) * 16

			screen.DrawImage(g.tilemapImg.SubImage(image.Rect(int(srcX), int(srcY), int(srcX+16), int(srcY+16))).(*ebiten.Image),
				&opts,
			)

			opts.GeoM.Reset()
		}
	}

	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.camera.X, g.camera.Y)

	screen.DrawImage(
		g.player.Img.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image),
		&opts,
	)

	opts.GeoM.Reset()

	for _, pokeball := range g.pokeball {
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(pokeball.X, pokeball.Y)
		opts.GeoM.Translate(g.camera.X, g.camera.Y)

		screen.DrawImage(
			pokeball.Img.SubImage(image.Rect(0, 0, 24, 24)).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}

	for _, collider := range g.colliders {
		vector.StrokeRect(
			screen,
			float32(collider.Min.X)+float32(g.camera.X),
			float32(collider.Min.Y)+float32(g.camera.Y),
			float32(collider.Dx()),
			float32(collider.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			true,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Pokemon Go")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	playerImg, _, err := ebitenutil.NewImageFromFile("../assets/images/Char_001.png")
	if err != nil {
		log.Fatal(err)
	}

	pokeballImg, _, err := ebitenutil.NewImageFromFile("../assets/images/pokeball.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapImg, _, err := ebitenutil.NewImageFromFile("../assets/images/Tileset.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("../assets/maps/pokeworld100.json")
	if err != nil {
		log.Fatal(err)
	}

	game := Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   0,
				Y:   0,
			},
		},

		pokeball: []*entities.Pokeball{
			{
				Img: pokeballImg,
				X:   100,
				Y:   100,
			},
			{
				Img: pokeballImg,
				X:   200,
				Y:   200,
			},
		},

		tilemapJSON: tilemapJSON,
		tilemapImg:  tilemapImg,
		camera:      NewCamera(0, 0),
		colliders: []image.Rectangle{
			image.Rect(100, 100, 124, 124),
		},
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
