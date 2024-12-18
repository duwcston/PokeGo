package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type Stats struct {
	HP         int `json:"HP"`
	Attack     int `json:"Attack"`
	Defense    int `json:"Defense"`
	Speed      int `json:"Speed"`
	Sp_Attack  int `json:"Sp_Attack"`
	Sp_Defense int `json:"Sp_Defense"`
}

type GenderRatio struct {
	MaleRatio   float32 `json:"MaleRatio"`
	FemaleRatio float32 `json:"FemaleRatio"`
}

type Profile struct {
	Height      float32     `json:"Height"`
	Weight      float32     `json:"Weight"`
	CatchRate   float32     `json:"CatchRate"`
	GenderRatio GenderRatio `json:"GenderRatio"`
	EggGroup    string      `json:"EggGroup"`
	HatchSteps  int         `json:"HatchSteps"`
	Abilities   string      `json:"Abilities"`
}

type DamegeWhenAttacked struct {
	Element     string  `json:"Element"`
	Coefficient float32 `json:"Coefficient"`
}

type Moves struct {
	Name        string `json:"Name"`
	Element     string `json:"Element"`
	Power       string `json:"Power"`
	Acc         int    `json:"Acc"`
	PP          int    `json:"PP"`
	Description string `json:"Description"`
}

type Pokemon struct {
	Name               string               `json:"Name"`
	Elements           []string             `json:"Elements"`
	Level              int                  `json:"Level"`
	EV                 float64              `json:"EV"`
	Stats              Stats                `json:"Stats"`
	Profile            Profile              `json:"Profile"`
	DamegeWhenAttacked []DamegeWhenAttacked `json:"DamegeWhenAttacked"`
	EvolutionLevel     int                  `json:"EvolutionLevel"`
	NextEvolution      string               `json:"NextEvolution"`
	Moves              []Moves              `json:"Moves"`
}

type Player struct {
	Name      string    `json:"Name"`
	Inventory []Pokemon `json:"Inventory"`
	Pokeball  int       `json:"Pokeball"`
}

const (
	pokedexPath = "../../Pokedex/pokedex.json"
	maxCapacity = 200
	HOST        = "localhost"
	PORT        = "3000"
)

// Track connected players
var (
	connectedPlayers = make(map[string]*net.UDPAddr)
	mutex            = sync.Mutex{}
)

func main() {
	// Resolve the server address
	addr, err := net.ResolveUDPAddr("udp4", HOST+":"+PORT)
	if err != nil {
		fmt.Println("Error resolving address: ", err)
		return
	}

	// Start the UDP server
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("UDP server listening on %s\n", HOST+":"+PORT)

	// Set up signal channel to catch termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Handle shutdown signals
	go func() {
		sig := <-sigs
		fmt.Println("Received signal:", sig)
		handleServerShutdown(conn)
	}()

	buffer := make([]byte, 1024)
	for {
		// Read incoming message
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}

		message := strings.TrimSpace(string(buffer[:n]))
		// fmt.Printf("Received '%s' from %s\n", message, clientAddr)

		// Handle player joining
		if strings.HasPrefix(message, "JOIN AS ") {
			playerName := strings.TrimPrefix(message, "JOIN AS ")
			handlePlayerConnection(playerName, conn, clientAddr)
			continue
		}

		if message == "QUIT" {
			playerName := findPlayerNameByAddr(clientAddr)
			handlePlayerDisconnection(playerName, conn, clientAddr)
			continue
		}

		if message == "gotcha" {
			sendRandomPokemon(conn, clientAddr)
			continue
		}

		conn.WriteToUDP([]byte("Unknown command\n"), clientAddr)
	}
}

func handlePlayerConnection(playerName string, conn *net.UDPConn, addr *net.UDPAddr) {
	mutex.Lock()
	connectedPlayers[playerName] = addr
	mutex.Unlock()

	fmt.Printf("Player '%s' connected from %s\n", playerName, addr)

	conn.WriteToUDP([]byte("Welcome "+playerName+"!\n"), addr)

	broadcastMessage(fmt.Sprintf("Player %s has joined the game!", playerName), addr, conn)

	player := &Player{
		Name:      playerName,
		Inventory: []Pokemon{},
		Pokeball:  10,
	}

	// Load player's inventory
	if err := LoadInventory(player); err != nil {
		fmt.Printf("Error loading inventory for %s: %v\n", playerName, err)
		conn.WriteToUDP([]byte("Failed to load inventory\n"), addr)
		return
	}
}

func handlePlayerDisconnection(playerName string, conn *net.UDPConn, addr *net.UDPAddr) {
	delete(connectedPlayers, addr.String())

	message := fmt.Sprintf("Player %s has left the game.", playerName)
	fmt.Println(message)
	broadcastMessage(message, addr, conn)

	conn.WriteToUDP([]byte("You have left the game.\n"), addr)
}

func handleServerShutdown(conn *net.UDPConn) {
	for _, addr := range connectedPlayers {
		conn.WriteToUDP([]byte("Server is shutting down\n"), addr)
	}

	conn.Close()
	os.Exit(0)
}

func broadcastMessage(message string, senderAddr *net.UDPAddr, conn *net.UDPConn) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, addr := range connectedPlayers {
		if addr.String() != senderAddr.String() {
			conn.WriteToUDP([]byte(message+"\n"), addr)
		}
	}
}

func sendRandomPokemon(conn *net.UDPConn, addr *net.UDPAddr) {
	// Read the Pokedex
	pokedex, err := os.ReadFile(pokedexPath)
	if err != nil {
		fmt.Println("Error reading Pokedex:", err)
		conn.WriteToUDP([]byte("Error reading Pokedex\n"), addr)
		return
	}

	// Unmarshal the Pokedex
	var pokedexJSON []Pokemon
	err = json.Unmarshal(pokedex, &pokedexJSON)
	if err != nil {
		fmt.Println("Error unmarshalling Pokedex:", err)
		conn.WriteToUDP([]byte("Error unmarshalling Pokedex\n"), addr)
		return
	}

	if len(pokedexJSON) == 0 {
		fmt.Println("Pokedex is empty")
		conn.WriteToUDP([]byte("Pokedex is empty\n"), addr)
		return
	}

	// Randomly select a Pokémon
	randomIndex := rand.Intn(len(pokedexJSON))
	randomPokemon := pokedexJSON[randomIndex]
	randomPokemon.EV = roundFloat(0.5+rand.Float64()*(0.5), 1)
	randomPokemon.Level = rand.Intn(10) + 1

	// Send Pokémon data to the client
	response := fmt.Sprintf("You caught a %s (Level %d, EV %.1f)\n", randomPokemon.Name, randomPokemon.Level, randomPokemon.EV)
	conn.WriteToUDP([]byte(response), addr)

	// Save the Pokémon to the player's inventory
	playerName := findPlayerNameByAddr(addr)
	mutex.Lock()
	for name, playerAddr := range connectedPlayers {
		if playerAddr.String() == addr.String() {
			player := &Player{
				Name:      name,
				Inventory: []Pokemon{randomPokemon},
			}
			player.Inventory = append(player.Inventory, randomPokemon)
			if err := SaveInventory(player); err != nil {
				fmt.Printf("Error saving inventory for %s: %v\n", playerName, err)
				conn.WriteToUDP([]byte("Failed to save inventory\n"), addr)
				mutex.Unlock()
				return
			}
		}
	}
	mutex.Unlock()

	broadcastMessage(fmt.Sprintf("Player %s caught a %s!", playerName, randomPokemon.Name), addr, conn)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func findPlayerNameByAddr(addr *net.UDPAddr) string {
	mutex.Lock()
	defer mutex.Unlock()

	for name, playerAddr := range connectedPlayers {
		if playerAddr.String() == addr.String() {
			return name
		}
	}

	return "Player not found"
}

func SaveInventory(player *Player) error {
	filename := filepath.Join("Inventories", fmt.Sprintf("Player_%s_Inventory.json", player.Name))
	data, err := json.MarshalIndent(player.Inventory, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling inventory: %v", err)
	}

	if err := os.WriteFile(filename, data, os.ModePerm); err != nil {
		return fmt.Errorf("error saving inventory to file: %v", err)
	}

	fmt.Printf("Player %s inventory saved to %s\n", player.Name, filename)
	return nil
}

func LoadInventory(player *Player) error {
	filename := filepath.Join("Inventories", fmt.Sprintf("Player_%s_Inventory.json", player.Name))

	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		player.Inventory = []Pokemon{}
		fmt.Printf("No inventory file found for Player %s. Initialized empty inventory.\n", player.Name)
		return nil
	} else if err != nil {
		return fmt.Errorf("error reading inventory file: %v", err)
	}

	if err := json.Unmarshal(data, &player.Inventory); err != nil {
		return fmt.Errorf("error unmarshalling inventory: %v", err)
	}

	fmt.Printf("Player %s inventory loaded from %s\n", player.Name, filename)
	return nil
}
