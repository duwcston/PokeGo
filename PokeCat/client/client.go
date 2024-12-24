package main

import (
	"bufio"
	"fmt"
	"image"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"PokeGo/constants"
	"PokeGo/entities"
)

const (
	HOST = "10.238.26.98"
	PORT = "3000"
	TYPE = "udp4" // Use UDP for communication
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

var conn *net.UDPConn
var serverAddr *net.UDPAddr

func main() {
	// Resolve the server address
	var err error
	serverAddr, err = net.ResolveUDPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		os.Exit(1)
	}

	// Dial the server using UDP
	conn, err = net.DialUDP(TYPE, nil, serverAddr)
	if err != nil {
		fmt.Println("Connection error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to listen for shutdown signals
	go func() {
		sig := <-sigs
		fmt.Println("Received signal:", sig)

		conn.Write([]byte("QUIT\n"))

		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	fmt.Println("Connected to server")
	fmt.Println("Enter 'QUIT' to leave the game")

	// Send player name to the server
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := scanner.Text()
	conn.Write([]byte("JOIN AS " + name + "\n"))

	// Start the game
	game := NewGame(conn)

	// Listen for messages from the server
	go listenToServer(conn, game)

	// Listen for player input
	go listenForInput(conn)

	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetWindowTitle(name + "'s Pokeworld")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game.spawnPokeballs()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func listenToServer(conn *net.UDPConn, game *Game) {
	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from server:", err)
			continue
		}
		message := string(buffer[:n])
		fmt.Println("Server:", message)

		// Check if the player has left pokestop
		if message == "Exited PokeStop.\n" {
			time.Sleep(2 * time.Second) // Add delay to prevent spamming
			game.isAccessedPokestop = false
		}

		if message == "You have left the game.\n" {
			os.Exit(0)
		}
	}
}

func listenForInput(conn *net.UDPConn) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()
		conn.Write([]byte(text + "\n"))
	}
}
