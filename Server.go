package main

import (
	"PokeGo/PokeBat"
	"PokeGo/model"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	HOST = "localhost"
	PORT = "8080"
	//TYPE = "udp"
)

var (
	clients     = make(map[string]*model.Player)
	clientsMu   sync.Mutex
	AllPokemons *[]model.Pokemon
)

func getAllPokemons(filename string) (*[]model.Pokemon, error) {
	// Read the JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a slice of Pokemon structs
	var pokemons []model.Pokemon
	err = json.Unmarshal(data, &pokemons)
	if err != nil {
		return nil, err
	}
	return &pokemons, nil
}
func getRandomPokemon(AllPokemons []model.Pokemon, nop int) (*[]model.Pokemon, error) {

	var randomPokemons []model.Pokemon

	// Generate a random index
	for range nop {
		index, err := PokeBat.RandomInt(int64(len(AllPokemons)))
		if err != nil {
			return nil, err
		}
		randomPokemons = append(randomPokemons, AllPokemons[index])
	}

	// Return the randomly selected Pokemon
	return &randomPokemons, nil

}

var IsInBattel = false

func main() {
	serverAddr := HOST + ":" + PORT

	// Resolve server address
	addr, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		fmt.Println("Error resolving address: ", err)
		return
	}
	AllPokemons, err = getAllPokemons("Pokedex/pokedex.json")
	if err != nil {
		fmt.Println("Error resolving address: ", err)
		return
	}
	// Listen on UDP
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Error listening: ", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	fmt.Println("Server listening on:", serverAddr)
	for {
		if !IsInBattel {
			//Receive message from client
			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading from UDP: ", err)
				continue
			}

			message := strings.TrimSpace(string(buffer[:n]))
			fmt.Println("Received: ", message, " from ", clientAddr)

			// Process commands

			handleCommand(conn, clientAddr, message)
		}
	}
}

func handleCommand(conn *net.UDPConn, addr *net.UDPAddr, message string) {

	parts := strings.SplitN(message, " ", 2)
	command := parts[0]
	switch command {
	case "LOGIN":
		if len(parts) < 2 {
			return
		}
		username := parts[1]
		registerPlayer(username, addr)
		conn.WriteToUDP([]byte(fmt.Sprintf("Wellcome %s: (@username to battle with other player)\n", username)), addr)
	case "LOGOUT":
		if len(parts) < 2 {
			return
		}
		username := parts[1]
		removePlayer(username)
		conn.WriteToUDP([]byte("Goodbye "+username), addr)
	case "BATTLE":
		if len(parts) < 2 {
			return
		}

		handleBat(conn, addr, parts[1])

	}

}

func registerPlayer(name string, addr *net.UDPAddr) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	//get random Pokemon
	pokemons, _ := getRandomPokemon(*AllPokemons, 5)
	clients[name] = &model.Player{Name: name, Addr: addr, Inventory: *pokemons}
	fmt.Println("Registered client: ", name, addr)
}

func removePlayer(name string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, name)
	fmt.Println("Removed client: ", name)
}

func handleBat(conn *net.UDPConn, senderAddr *net.UDPAddr, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	var Sender *model.Player
	for _, client := range clients {
		if client.Addr.String() == senderAddr.String() {
			Sender = client
			break
		}
	}

	// Check for @<username> or @all command
	if strings.HasPrefix(message, "@") {
		parts := strings.SplitN(message, " ", 2)
		if len(parts) < 2 {
			return
		}
		target := parts[0][1:] // Remove "@" prefix
		msg := parts[1]

		broadcastMsg := fmt.Sprintf("Broadcast : Battel between %s and %s", Sender.Name, target)
		for _, client := range clients {
			if client.Addr.String() != senderAddr.String() { // Exclude sender
				conn.WriteToUDP([]byte(broadcastMsg), client.Addr)
			}
		}
		// Private message to specific user
		if Receiver, ok := clients[target]; ok {
			privateMsg := fmt.Sprintf("Bat from %s: %s", Sender.Name, msg)
			conn.WriteToUDP([]byte(privateMsg), Receiver.Addr)
			IsInBattel = true
			winner, LevelUpPokemons := PokeBat.Battle(Sender, Receiver, conn, Sender.Addr, Receiver.Addr)
			IsInBattel = false
			// Evolution
			PokeBat.EvolutionProcess(*clients[winner], LevelUpPokemons, *AllPokemons, conn)
		} else {
			conn.WriteToUDP([]byte("User "+target+" not found"), senderAddr)
		}

		broadcastMsg = fmt.Sprintf("Broadcast : Battle between %s and %s over", Sender.Name, target)
		for _, client := range clients {
			if client.Addr.String() != senderAddr.String() { // Exclude sender
				conn.WriteToUDP([]byte(broadcastMsg), client.Addr)
			}
		}
	}
}
