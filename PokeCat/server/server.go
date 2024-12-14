package main

import (
	"PokeGo/entities"
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
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
	EV                 int                  `json:"EV"`
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
	ID        string    `json:"ID"`
	X         float64   `json:"X"`
	Y         float64   `json:"Y"`
	Inventory []Pokemon `json:"Inventory"`
	Addr      *net.UDPAddr
	sync.Mutex
}

const (
	pokedexPath = "../../Pokedex/pokedex.json"
)

var (
	clients = make(map[net.Conn]bool)
	lock    sync.Mutex
	stateMu sync.Mutex
)

type Game struct {
	otherPlayers map[string]*entities.Player
}

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
	playerID := fmt.Sprintf("Player%d", rand.Intn(1000))

	defer func() {
		lock.Lock()
		stateMu.Lock()
		delete(clients, conn)
		lock.Unlock()
		stateMu.Unlock()
		broadcast(fmt.Sprintf("DISCONNECT %s", playerID), nil)
		conn.Close()
	}()

	broadcast(fmt.Sprintf("NEWPLAYER %s", playerID), conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Println("Received:", msg)

		if msg == "gotcha" {
			sendRandomPokemon(conn)
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

func sendRandomPokemon(conn net.Conn) {
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

	// Serialize the Pokémon as JSON
	pokemonData, err := json.MarshalIndent(randomPokemon.Name, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling Pokémon:", err)
		conn.Write([]byte("Error marshalling Pokémon\n"))
		return
	}

	// Send the Pokémon data to the client
	conn.Write(append(pokemonData, '\n'))
	fmt.Println("Sent Pokémon:", randomPokemon.Name)
}
