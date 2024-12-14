package main

import (
	"PokeGo/animations"
	"PokeGo/constants"
	"PokeGo/entities"
	"PokeGo/spritesheet"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	// "github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	player            *entities.Player
	playerSpriteSheet *spritesheet.SpriteSheet
	pokeball          []*entities.Pokeball
	tilemapJSON       *TilemapJSON
	tilemapImg        *ebiten.Image
	camera            *Camera
	colliders         []image.Rectangle
}

func NewGame() *Game {
	playerImg, _, err := ebitenutil.NewImageFromFile("../../assets/images/Char_001.png")
	if err != nil {
		log.Fatal(err)
	}

	pokeballImg, _, err := ebitenutil.NewImageFromFile("../../assets/images/pokeball.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapImg, _, err := ebitenutil.NewImageFromFile("../../assets/images/Tileset.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("../../assets/maps/pokeworld100.json")
	if err != nil {
		log.Fatal(err)
	}

	spritesheet := spritesheet.NewSpriteSheet(4, 4, constants.SpriteSize)

	return &Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   rand.Float64() * float64(constants.Tilesize*tilemapJSON.Layers[0].Width),
				Y:   rand.Float64() * float64(constants.Tilesize*tilemapJSON.Layers[0].Height),
			},
			Animations: map[entities.PlayerState]*animations.Animation{
				entities.Down:  animations.NewAnimation(0, 3, 1, 10.0),
				entities.Left:  animations.NewAnimation(4, 7, 1, 10.0),
				entities.Right: animations.NewAnimation(8, 11, 1, 10.0),
				entities.Up:    animations.NewAnimation(12, 15, 1, 10.0),
			},
		},

		playerSpriteSheet: spritesheet,

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
			{
				Min: image.Point{0, 0},
				Max: image.Point{constants.Tilesize * tilemapJSON.Layers[0].Width, constants.Tilesize},
			},
			{
				Min: image.Point{0, 0},
				Max: image.Point{constants.Tilesize, constants.Tilesize * tilemapJSON.Layers[0].Height},
			},
			{
				Min: image.Point{0, constants.Tilesize*tilemapJSON.Layers[0].Height - constants.Tilesize},
				Max: image.Point{constants.Tilesize * tilemapJSON.Layers[0].Width, constants.Tilesize * tilemapJSON.Layers[0].Height},
			},
			{
				Min: image.Point{constants.Tilesize*tilemapJSON.Layers[0].Width - constants.Tilesize, 0},
				Max: image.Point{constants.Tilesize * tilemapJSON.Layers[0].Width, constants.Tilesize * tilemapJSON.Layers[0].Height},
			},
		},
	}
}

func (g *Game) Update() error {
	// fmt.Println("Player X:", int(g.player.X), "Player Y:", int(g.player.Y))

	g.player.Dx = 0
	g.player.Dy = 0

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -2
		g.player.Dy = 0
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 2
		g.player.Dy = 0
	} else if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dx = 0
		g.player.Dy = -2
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dx = 0
		g.player.Dy = 2
	}

	g.player.X += g.player.Dx
	g.player.Y += g.player.Dy

	CheckCollisionHorizontal(g.player.Sprite, g.colliders)
	CheckCollisionVertical(g.player.Sprite, g.colliders)

	activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activeAnim != nil {
		activeAnim.Update()
	}

	for i := len(g.pokeball) - 1; i >= 0; i-- {
		pokeball := g.pokeball[i]
		if (int(g.player.X) >= int(pokeball.X)-10 && int(g.player.X) <= int(pokeball.X)+10) &&
			(int(g.player.Y) >= int(pokeball.Y)-10 && int(g.player.Y) <= int(pokeball.Y)+10) {
			fmt.Println("Pokeball collected!")
			conn.Write([]byte("gotcha\n"))
			g.pokeball = append(g.pokeball[:i], g.pokeball[i+1:]...)
		}
	}

	// Testing sending message to server
	// if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
	// 	fmt.Println("Sending message to server...")
	// 	conn.Write([]byte("gotcha\n"))
	// }

	g.camera.FollowPlayer(g.player.X+constants.Tilesize/2, g.player.Y+constants.Tilesize/2, constants.ScreenWidth, constants.ScreenHeight)
	g.camera.Constrain(
		float64(g.tilemapJSON.Layers[0].Width)*constants.Tilesize,
		float64(g.tilemapJSON.Layers[0].Height)*constants.Tilesize,
		constants.ScreenWidth, constants.ScreenHeight,
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

			tileX := float64((index % layer.Width) * constants.Tilesize)
			tileY := float64((index / layer.Width) * constants.Tilesize)

			opts.GeoM.Translate(tileX, tileY)
			opts.GeoM.Translate(g.camera.X, g.camera.Y)

			srcX := float64((tileID-1)%64) * constants.Tilesize
			srcY := float64((tileID-1)/64) * constants.Tilesize

			screen.DrawImage(g.tilemapImg.SubImage(image.Rect(int(srcX), int(srcY), int(srcX+constants.Tilesize), int(srcY+constants.Tilesize))).(*ebiten.Image),
				&opts,
			)

			opts.GeoM.Reset()
		}
	}

	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.camera.X, g.camera.Y)

	playerFrame := 0
	activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activeAnim != nil {
		playerFrame = activeAnim.Frame()
	}

	screen.DrawImage(
		g.player.Img.SubImage(g.playerSpriteSheet.Rect(playerFrame)).(*ebiten.Image),
		&opts,
	)

	opts.GeoM.Reset()

	for _, pokeball := range g.pokeball {
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(pokeball.X, pokeball.Y)
		opts.GeoM.Translate(g.camera.X, g.camera.Y)

		screen.DrawImage(
			pokeball.Img.SubImage(image.Rect(0, 0, constants.SpriteSize, constants.SpriteSize)).(*ebiten.Image),
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
	return constants.ScreenWidth, constants.ScreenHeight
}
