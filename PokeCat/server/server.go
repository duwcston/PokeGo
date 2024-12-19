package main

import (
	"PokeGo/PokeBat"
	"PokeGo/model"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	pokedexPath = "../../Pokedex/pokedex.json"
	//pokedexPath = "Pokedex/pokedex.json"

	maxCapacity = 200
	HOST        = "localhost"
	PORT        = "3000"
	//InventoryPath = "PokeCat/Inventories/Player_%s_Inventory.json"
	InventoryPath = "../Inventories/Player_%s_Inventory.json"
)

// Track connected players
var (
	connectedPlayers = make(map[string]*model.Player)
	mutex            = sync.Mutex{}
	IsInBattel       = false
	AllPokemons      *[]model.Pokemon
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
	AllPokemons, err = getAllPokemons(pokedexPath)
	if err != nil {
		fmt.Println("Error resolving address: ", err)
		return
	}
	buffer := make([]byte, 1024)
	for {
		if !IsInBattel {
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

			var Sender *model.Player
			for _, client := range connectedPlayers {
				if client.Addr.String() == clientAddr.String() {
					Sender = client
					break
				}
			}
			if message == "QUIT" {
				playerName := findPlayerNameByAddr(clientAddr)
				handlePlayerDisconnection(playerName, conn, clientAddr)
				continue
			}

			if message == "gotcha" {
				sendRandomPokemon(*AllPokemons, conn, clientAddr)
				continue
			}
			if message == "Inventory" {
				for _, inv := range Sender.Inventory {
					inventoryDetails := fmt.Sprintf("Player Inventory: Name: %s, Level: %d, HP: %d", inv.Name, inv.Level, inv.Stats.HP)
					_, err := conn.WriteToUDP([]byte(inventoryDetails), Sender.Addr)
					if err != nil {
						fmt.Println("Error sending connect message to client:", err)
					}
				}
				continue
			}
			if strings.HasPrefix(message, "@") {
				msg := ""
				parts := strings.SplitN(message, " ", 2)
				if len(parts) >= 2 {
					msg = parts[1]
				}
				target := parts[0][1:] // Remove "@" prefix

				broadcastMsg := fmt.Sprintf("Broadcast : Battle between %s and %s", Sender.Name, target)
				for _, client := range connectedPlayers {
					if client.Addr.String() != Sender.Addr.String() { // Exclude sender
						conn.WriteToUDP([]byte(broadcastMsg), client.Addr)
					}
				}
				// Private message to specific user
				if Receiver, ok := connectedPlayers[target]; ok {
					privateMsg := fmt.Sprintf("Bat from %s: %s", Sender.Name, msg)
					conn.WriteToUDP([]byte(privateMsg), Receiver.Addr)
					IsInBattel = true
					SenderResult, ReceiverResult := PokeBat.Battle(Sender, Receiver, *AllPokemons, conn, Sender.Addr, Receiver.Addr)
					IsInBattel = false
					connectedPlayers[SenderResult.Name].Inventory = SenderResult.Inventory
					connectedPlayers[ReceiverResult.Name].Inventory = ReceiverResult.Inventory
					if err := SaveInventory(Sender); err != nil {
						fmt.Printf("Error saving inventory for %s: %v\n", Sender.Name, err)
						conn.WriteToUDP([]byte("Failed to save inventory\n"), Sender.Addr)
						mutex.Unlock()
						return
					}
					if err := SaveInventory(Receiver); err != nil {
						fmt.Printf("Error saving inventory for %s: %v\n", Sender.Name, err)
						conn.WriteToUDP([]byte("Failed to save inventory\n"), Sender.Addr)
						mutex.Unlock()
						return
					}

					// Evolution
					//if LevelUpPokemons != nil {
					//	PokeBat.EvolutionProcess(*connectedPlayers[winner], LevelUpPokemons, *AllPokemons, conn)
					//	if err := SaveInventory(Sender); err != nil {
					//		fmt.Printf("Error saving inventory for %s: %v\n", Sender.Name, err)
					//		conn.WriteToUDP([]byte("Failed to save inventory\n"), Sender.Addr)
					//		mutex.Unlock()
					//		return
					//	}
					//}
					continue
				} else {
					conn.WriteToUDP([]byte("User "+target+" not found"), Sender.Addr)
					continue
				}
			}

			conn.WriteToUDP([]byte("Unknown command\n"), clientAddr)
		}
	}
}

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

func handlePlayerConnection(playerName string, conn *net.UDPConn, addr *net.UDPAddr) {
	fmt.Printf("Player '%s' connected from %s\n", playerName, addr)

	conn.WriteToUDP([]byte("Welcome "+playerName+"!\n"), addr)
	conn.WriteToUDP([]byte("\n Enter 'Inventory' to view Pokemons \n Enter '@username' to battle"), addr)

	broadcastMessage(fmt.Sprintf("Player %s has joined the game!", playerName), addr, conn)

	player := &model.Player{
		Name:      playerName,
		Addr:      addr,
		Inventory: []model.Pokemon{},
	}
	connectedPlayers[playerName] = player
	// Load player's inventory
	if err := LoadInventory(player, addr); err != nil {
		fmt.Printf("Error loading inventory for %s: %v\n", playerName, err)
		conn.WriteToUDP([]byte("Failed to load inventory\n"), addr)
		return
	}
}

func handlePlayerDisconnection(playerName string, conn *net.UDPConn, addr *net.UDPAddr) {
	delete(connectedPlayers, playerName)

	message := fmt.Sprintf("Player %s has left the game.", playerName)
	fmt.Println(message)
	broadcastMessage(message, addr, conn)

	conn.WriteToUDP([]byte("You have left the game.\n"), addr)
}

func handleServerShutdown(conn *net.UDPConn) {
	for _, player := range connectedPlayers {
		conn.WriteToUDP([]byte("Server is shutting down\n"), player.Addr)
	}

	conn.Close()
	os.Exit(0)
}

func broadcastMessage(message string, senderAddr *net.UDPAddr, conn *net.UDPConn) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, player := range connectedPlayers {
		if player.Addr.String() != senderAddr.String() {
			conn.WriteToUDP([]byte(message+"\n"), player.Addr)
		}
	}
}

func sendRandomPokemon(pokedexJSON []model.Pokemon, conn *net.UDPConn, addr *net.UDPAddr) {

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
	for _, player := range connectedPlayers {
		if player.Addr.String() == addr.String() {
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
func CreatePlayertoJson(filename string, player *model.Player) {
	updatedData, err := json.MarshalIndent(player, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Create New Player Inventory")
	_, err = os.Create(filename)
	if err != nil {
		fmt.Println("Error...:", err)
	}
	if err = os.WriteFile(filename, updatedData, 0666); err != nil {
		fmt.Println("Error ...... :", err)
	}

}
func findPlayerNameByAddr(addr *net.UDPAddr) string {
	mutex.Lock()
	defer mutex.Unlock()

	for name, player := range connectedPlayers {
		if player.Addr.String() == addr.String() {
			return name
		}
	}

	return "Player not found"
}

func SaveInventory(player *model.Player) error {
	//filename := filepath.Join("Inventories", fmt.Sprintf("Player_%s_Inventory.json", player.Name))
	filename := fmt.Sprintf(InventoryPath, player.Name)
	data, err := json.MarshalIndent(player, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling inventory: %v", err)
	}

	if err := os.WriteFile(filename, data, os.ModePerm); err != nil {
		return fmt.Errorf("error saving inventory to file: %v", err)
	}

	fmt.Printf("Player %s inventory saved to %s\n", player.Name, filename)
	return nil
}

func LoadInventory(player *model.Player, addr *net.UDPAddr) error {
	//filename := filepath.Join("Inventories", fmt.Sprintf("Player_%s_Inventory.json", player.Name))
	filename := fmt.Sprintf(InventoryPath, player.Name)
	data, err := os.ReadFile(filename)
	demoaddr := addr.Port
	if os.IsNotExist(err) {
		player.Inventory = []model.Pokemon{}
		fmt.Printf("No inventory file found for Player %s. Initialized empty inventory.\n", player.Name)
		//Create Json File if new player
		CreatePlayertoJson(filename, player)
		return nil
	} else if err != nil {
		return fmt.Errorf("error reading inventory file: %v", err)
	}
	//If old player load data
	if err := json.Unmarshal(data, &player); err != nil {
		return fmt.Errorf("error unmarshalling inventory: %v", err)
	}
	fmt.Printf("Player %s inventory loaded from %s\n", player.Name, filename)
	player.Addr.Port = demoaddr
	return nil
}
