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
	"net"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	player            *entities.Player
	playerSpriteSheet *spritesheet.SpriteSheet
	pokeball          []*entities.Pokeball
	tilemapJSON       *TilemapJSON
	tilemapImg        *ebiten.Image
	camera            *Camera
	colliders         []image.Rectangle

	// Multiplayer simulation
	otherPlayers map[string]*entities.Player
	serverConn   net.Conn

	// Timer for spawning Pokeballs
	spawnInterval     float64
	maxPokeballs      int
	lastPokeballSpawn time.Time

	// Auto-move
	isAutoMoveEnabled bool
	moveTimer         time.Duration
	lastMoveTime      time.Time

	// Access pokestop
	isAccessedPokestop bool
}

var (
	pokeballImg, _, _ = ebitenutil.NewImageFromFile("../../assets/images/pokeball.png")
)

func NewGame(conn *net.UDPConn) *Game {
	playerImg, _, err := ebitenutil.NewImageFromFile("../../assets/images/Char_001.png")
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

			Img: playerImg,
			X:   rand.Float64() * float64(constants.Tilesize*tilemapJSON.Layers[0].Width-constants.Tilesize),
			Y:   rand.Float64() * float64(constants.Tilesize*tilemapJSON.Layers[0].Height-constants.Tilesize),

			Animations: map[entities.PlayerState]*animations.Animation{
				entities.Down:  animations.NewAnimation(0, 3, 1, 10.0),
				entities.Left:  animations.NewAnimation(4, 7, 1, 10.0),
				entities.Right: animations.NewAnimation(8, 11, 1, 10.0),
				entities.Up:    animations.NewAnimation(12, 15, 1, 10.0),
			},
		},

		playerSpriteSheet: spritesheet,

		pokeball: []*entities.Pokeball{},

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
			{
				Min: image.Point{constants.Tilesize * 18, constants.Tilesize * 17},
				Max: image.Point{constants.Tilesize * 23, constants.Tilesize * 22},
			},
			{
				Min: image.Point{constants.Tilesize * 18, constants.Tilesize * 76},
				Max: image.Point{constants.Tilesize * 23, constants.Tilesize * 82},
			},
			{
				Min: image.Point{constants.Tilesize * 78, constants.Tilesize * 17},
				Max: image.Point{constants.Tilesize * 83, constants.Tilesize * 22},
			},
			{
				Min: image.Point{constants.Tilesize * 78, constants.Tilesize * 76},
				Max: image.Point{constants.Tilesize * 83, constants.Tilesize * 81},
			},
		},

		serverConn:   conn,
		otherPlayers: map[string]*entities.Player{},

		spawnInterval:     5,  // 5 minutes in seconds
		maxPokeballs:      50, // Spawn 50 Pokeballs
		lastPokeballSpawn: time.Now(),

		isAutoMoveEnabled: false,
		moveTimer:         time.Second, // Move every second
		lastMoveTime:      time.Now(),  // Track the last move time

		isAccessedPokestop: false,
	}
}

func (g *Game) Update() error {
	// fmt.Println("Player X:", int(g.player.X), "Player Y:", int(g.player.Y))
	if g.isAccessedPokestop {
		return nil
	}

	if g.isAutoMoveEnabled && time.Since(g.lastMoveTime) >= g.moveTimer {
		randomDirection := rand.Intn(4)
		switch randomDirection {
		case 0:
			g.player.Dx = -constants.Tilesize * 4
			g.player.Dy = 0
		case 1:
			g.player.Dx = constants.Tilesize * 4
			g.player.Dy = 0
		case 2:
			g.player.Dx = 0
			g.player.Dy = -constants.Tilesize * 4
		case 3:
			g.player.Dx = 0
			g.player.Dy = constants.Tilesize * 4
		}
		g.player.X += g.player.Dx
		g.player.Y += g.player.Dy
		g.lastMoveTime = time.Now()
	}

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

	CheckCollisionHorizontal(g.player, g.colliders)
	CheckCollisionVertical(g.player, g.colliders)

	activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activeAnim != nil {
		activeAnim.Update()
	}

	if time.Since(g.lastPokeballSpawn) >= time.Duration(g.spawnInterval)*time.Minute {
		g.spawnPokeballs()
		g.lastPokeballSpawn = time.Now() // reset the timer
	}

	for i := len(g.pokeball) - 1; i >= 0; i-- {
		pokeball := g.pokeball[i]
		if (int(g.player.X) >= int(pokeball.X)-10 && int(g.player.X) <= int(pokeball.X)+10) &&
			(int(g.player.Y) >= int(pokeball.Y)-10 && int(g.player.Y) <= int(pokeball.Y)+10) {
			fmt.Println("Try to catch a pokemon...")
			g.serverConn.Write([]byte("gotcha\n"))
			g.pokeball = append(g.pokeball[:i], g.pokeball[i+1:]...)
		}
	}

	// Access pokestop to get pokeballs and berries
	if !g.isAccessedPokestop && !g.isAutoMoveEnabled && ((int(g.player.X) >= constants.Tilesize*20-1 && int(g.player.X) <= constants.Tilesize*20+1) &&
		(int(g.player.Y) >= constants.Tilesize*22-1 && int(g.player.Y) <= constants.Tilesize*22+1) ||
		(int(g.player.X) >= constants.Tilesize*20-1 && int(g.player.X) <= constants.Tilesize*20+1) &&
			(int(g.player.Y) >= constants.Tilesize*82-1 && int(g.player.Y) <= constants.Tilesize*82+1) ||
		(int(g.player.X) >= constants.Tilesize*80-1 && int(g.player.X) <= constants.Tilesize*80+1) &&
			(int(g.player.Y) >= constants.Tilesize*22-1 && int(g.player.Y) <= constants.Tilesize*22+1) ||
		(int(g.player.X) >= constants.Tilesize*79-2 && int(g.player.X) <= constants.Tilesize*79+2) &&
			(int(g.player.Y) >= constants.Tilesize*81-2 && int(g.player.Y) <= constants.Tilesize*81+2)) {
		fmt.Println("Pokestop accessed!")
		g.serverConn.Write([]byte("pokestop\n"))
		g.isAccessedPokestop = true
	}

	g.camera.FollowPlayer(g.player.X+constants.Tilesize/2, g.player.Y+constants.Tilesize/2, constants.ScreenWidth, constants.ScreenHeight)
	g.camera.Constrain(
		float64(g.tilemapJSON.Layers[0].Width)*constants.Tilesize,
		float64(g.tilemapJSON.Layers[0].Height)*constants.Tilesize,
		constants.ScreenWidth, constants.ScreenHeight,
	)

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.isAutoMoveEnabled = !g.isAutoMoveEnabled
		fmt.Println("Auto-move enabled:", g.isAutoMoveEnabled)
	}

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

	// Draw other players
	for _, otherPlayer := range g.otherPlayers {
		opts.GeoM.Translate(otherPlayer.X, otherPlayer.Y)
		opts.GeoM.Translate(g.camera.X, g.camera.Y)

		playerFrame := 0
		activeAnim := otherPlayer.ActiveAnimation(int(otherPlayer.Dx), int(otherPlayer.Dy))
		if activeAnim != nil {
			playerFrame = activeAnim.Frame()
		}

		screen.DrawImage(
			otherPlayer.Img.SubImage(g.playerSpriteSheet.Rect(playerFrame)).(*ebiten.Image),
			&opts,
		)

		opts.GeoM.Reset()
	}

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

	// for _, collider := range g.colliders {
	// 	vector.StrokeRect(
	// 		screen,
	// 		float32(collider.Min.X)+float32(g.camera.X),
	// 		float32(collider.Min.Y)+float32(g.camera.Y),
	// 		float32(collider.Dx()),
	// 		float32(collider.Dy()),
	// 		1.0,
	// 		color.RGBA{255, 0, 0, 255},
	// 		true,
	// 	)
	// }

	ebitenutil.DebugPrint(screen, "Auto mode: "+fmt.Sprint(func() string {
		if g.isAutoMoveEnabled {
			return "ON"
		}
		return "OFF"
	}()))

	ebitenutil.DebugPrint(screen, "\nisAccessedPokestop: "+fmt.Sprint(g.isAccessedPokestop)) // Debug print for accessing pokestop
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

func (g *Game) spawnPokeballs() {
	// Clear the previous Pokéballs before spawning new ones
	g.pokeball = []*entities.Pokeball{}

	// Spawn 50 random pokeballs
	for i := 0; i < g.maxPokeballs; i++ {
		randomX := rand.Float64() * float64(constants.Tilesize*g.tilemapJSON.Layers[0].Width-constants.Tilesize)
		randomY := rand.Float64() * float64(constants.Tilesize*g.tilemapJSON.Layers[0].Height-constants.Tilesize)

		// Add new pokeball to the list
		g.pokeball = append(g.pokeball, &entities.Pokeball{
			Img: pokeballImg,
			X:   randomX,
			Y:   randomY,
		})
	}

	fmt.Println("50 Pokéballs spawned on the map!")
}

func (g *Game) ToggleAutoMove() {
	g.isAutoMoveEnabled = !g.isAutoMoveEnabled
	if g.isAutoMoveEnabled {
		g.lastMoveTime = time.Now() // Reset the move timer when enabling auto-move
	}
}
