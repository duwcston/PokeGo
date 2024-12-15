package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

type Pokedex struct {
	Pokemon     Pokemon `json:"Pokemon"`
	CoordinateX int
	CoordinateY int
}

type Player struct {
	Name      string    `json:"Name"`
	Inventory []Pokemon `json:"Inventory"`
}

const (
	pokedexPath = "../../Pokedex/pokedex.json"
	maxCapacity = 200
)

var (
	clients = make(map[net.Conn]bool)
	lock    sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on port 3000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		lock.Lock()
		clients[conn] = true
		lock.Unlock()
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Read the player's ID from the connection
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	playerID := scanner.Text()
	if strings.HasPrefix(playerID, "JOIN AS ") {
		playerID = strings.TrimPrefix(playerID, "JOIN AS ")
	} else {
		fmt.Println("Invalid JOIN AS message:", playerID)
	}
	// conn.Close()

	player := &Player{
		Name:      playerID,
		Inventory: []Pokemon{},
	}

	// Load the player's inventory from the JSON file
	if err := LoadInventory(player); err != nil {
		fmt.Printf("Error loading inventory for Player %s: %v\n", playerID, err)
	}

	lock.Lock()
	clients[conn] = true
	lock.Unlock()

	defer func() {
		lock.Lock()
		delete(clients, conn)
		lock.Unlock()

		// Save the player's inventory to the JSON file
		if err := SaveInventory(player); err != nil {
			fmt.Printf("Error saving inventory for Player %s: %v\n", playerID, err)
		}

		broadcast(fmt.Sprintf("DISCONNECT %s", playerID), nil)
		conn.Close()
	}()

	// scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Println("Received:", msg)

		broadcast(fmt.Sprintf("NEWPLAYER %s", playerID), conn)

		if msg == "gotcha" {
			sendRandomPokemon(conn, player)
		} else {
			broadcast(msg, conn)
		}
	}
}

func broadcast(msg string, sender net.Conn) {
	lock.Lock()
	defer lock.Unlock()
	for client := range clients {
		if client != sender {
			client.Write([]byte(msg + "\n"))
		}
	}
}

func sendRandomPokemon(conn net.Conn, player *Player) {
	// Read the Pokedex
	pokedex, err := os.ReadFile(pokedexPath)
	if err != nil {
		fmt.Println("Error reading Pokedex:", err)
		conn.Write([]byte("Error reading Pokedex\n"))
		return
	}

	// Unmarshal the Pokedex
	var pokedexJSON []Pokemon
	err = json.Unmarshal(pokedex, &pokedexJSON)
	if err != nil {
		fmt.Println("Error unmarshalling Pokedex:", err)
		conn.Write([]byte("Error unmarshalling Pokedex\n"))
		return
	}

	// Check if the Pokedex contains entries
	if len(pokedexJSON) == 0 {
		fmt.Println("Pokedex is empty")
		conn.Write([]byte("Pokedex is empty\n"))
		return
	}

	// Randomly select a Pokémon
	randomIndex := rand.Intn(len(pokedexJSON))
	randomPokemon := pokedexJSON[randomIndex]
	randomPokemon.EV = roundFloat(0.5+rand.Float64()*(0.5), 1)
	randomPokemon.Level = rand.Intn(10) + 1

	// Add to the player's inventory if not full
	if len(player.Inventory) < maxCapacity {
		player.Inventory = append(player.Inventory, randomPokemon)
		fmt.Printf("Player %s captured Pokémon %s (Level %d, EV %.1f)\n", player.Name, randomPokemon.Name, randomPokemon.Level, randomPokemon.EV)
	} else {
		fmt.Printf("Player %s's inventory is full. Cannot capture Pokémon %s\n", player.Name, randomPokemon.Name)
		conn.Write([]byte("Inventory full\n"))
		return
	}

	// Send the Pokémon data to the client
	conn.Write([]byte(player.Name + " captured a " + randomPokemon.Name + "\n"))
	fmt.Println("Sent Pokémon:", randomPokemon.Name, randomPokemon.Level, randomPokemon.EV)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
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

// LoadInventory loads a player's inventory from a JSON file.
func LoadInventory(player *Player) error {
	filename := filepath.Join("Inventories", fmt.Sprintf("Player_%s_Inventory.json", player.Name))

	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		// No inventory file exists, initialize an empty inventory
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
