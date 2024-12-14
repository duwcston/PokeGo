package PokeBat

import (
	"PokeGo/model"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"strconv"
	"strings"
)

func PlayerMove(player1Pokemon, player2Pokemon *model.Pokemon, player1Pokemons *[]model.Pokemon, PlayerName string, conn *net.UDPConn, addr1, addr2 *net.UDPAddr) (*model.Pokemon, bool) {
	if !isAlive(player1Pokemon) {
		fmt.Println(player1Pokemon.Name, "is dead")
		conn.WriteToUDP([]byte(fmt.Sprintf("%s is dead\n", player1Pokemon.Name)), addr1)
		player1Pokemon = switchPokemon(*player1Pokemons, conn, addr1)
		if player1Pokemon == nil {
			fmt.Println(PlayerName + "has no pokemon left")
			fmt.Println(PlayerName + " lost")
			conn.WriteToUDP([]byte("You have no pokemon left. You lost.\n"), addr1)
			conn.WriteToUDP([]byte("Enemy has no pokemon left. You wins.\n"), addr2)
			return nil, true
		} else {
			fmt.Printf("%s switched to %s \n", PlayerName, player1Pokemon.Name)
			conn.WriteToUDP([]byte(fmt.Sprintf("You switched to %s\n", player1Pokemon.Name)), addr1)
		}
	}

	fmt.Printf("%s turn. Your current pokemon is %s. Choose your action(attack,switch,surrender,?):\n", PlayerName, player1Pokemon.Name)
	conn.WriteToUDP([]byte(fmt.Sprintf("Your turn. Your current pokemon is %s. Choose your action(attack,switch,surrender,?):\n", player1Pokemon.Name)), addr1)
	command := readCommands(conn, addr1)
	switch command {
	case "attack":
		attack(player1Pokemon, player2Pokemon, conn, addr1)
	case "switch":
		displaySelectedPokemons(*player1Pokemons, conn, addr1)
		player1Pokemon = switchToChosenPokemon(*player1Pokemons, conn, addr1)
		fmt.Printf("%s switched to %s \n", PlayerName, player1Pokemon.Name)
		conn.WriteToUDP([]byte(fmt.Sprintf("Switched to %s\n", player1Pokemon.Name)), addr1)
	case "surrender":
		fmt.Printf("%s lost\n", PlayerName)
		conn.WriteToUDP([]byte("You surrender to Enemy. You lost.\n"), addr1)
		conn.WriteToUDP([]byte("Enemy surrenders. You wins.\n"), addr2)
		return player1Pokemon, true
		break
	case "?":
		displayCommandsList(conn, addr1)
	}
	return player1Pokemon, false
}

func attack(attacker *model.Pokemon, defender *model.Pokemon, conn *net.UDPConn, addr *net.UDPAddr) {
	// Calculate the damage
	var dmg float32
	var attackerMove = chooseAttack()
	fmt.Println(attacker.Name, "chose", attackerMove, "to attack", defender.Name)
	conn.WriteToUDP([]byte(fmt.Sprintf("%s chose %s to attack %s\n", attacker.Name, attackerMove, defender.Name)), addr)

	switch attackerMove {
	case "Tackle":
		dmg = float32(attacker.Stats.Attack - defender.Stats.Defense)
	case "Special":
		attackingElement := attacker.Elements
		dmgWhenAttacked := defender.DamegeWhenAttacked
		defendingElement := []string{}
		for _, element := range dmgWhenAttacked {
			defendingElement = append(defendingElement, element.Element)
		}
		highestCoefficient := float32(0)

		// Check for the highest coefficient
		for i, element := range defendingElement {
			if isContain(attackingElement, element) {
				if highestCoefficient < dmgWhenAttacked[i].Coefficient {
					highestCoefficient = dmgWhenAttacked[i].Coefficient
				}
			}
		}

		// If the attacker has an element that the defender doesn't have, set the coefficient to 1
		for _, element := range defendingElement {
			if !isContain(attackingElement, element) && highestCoefficient < 1 {
				highestCoefficient = 1
			}
		}

		dmg = float32(attacker.Stats.Sp_Attack)*highestCoefficient - float32(defender.Stats.Sp_Defense)
	}

	if dmg < 0 {
		dmg = 0
	}

	defender.Stats.HP -= int(dmg)
	fmt.Println(attacker.Name, "attacked", defender.Name, "with", attackerMove, "and dealt", dmg, "damage", defender.Stats.HP, "HP left")
	conn.WriteToUDP([]byte(fmt.Sprintf("%s attacked %s with %s and dealt %.2f damage %d HP left\n", attacker.Name, defender.Name, attackerMove, dmg, defender.Stats.HP)), addr)
}

func chooseAttack() string {
	n, _ := RandomInt(2)
	if n == 1 {
		return "Tackle"
	}
	return "Special"
}

func isContain[T any](arr []T, element T) bool {
	for _, a := range arr {
		if reflect.DeepEqual(a, element) {
			return true
		}
	}
	return false
}

func getHighestSpeedPokemon(Pokemons []model.Pokemon) *model.Pokemon {
	var highestSpeed = 0
	var choosenPokemonIndex = 0
	for i, pokemon := range Pokemons {
		if pokemon.Stats.Speed > highestSpeed {
			highestSpeed = pokemon.Stats.Speed
			choosenPokemonIndex = i
		}
	}

	return &Pokemons[choosenPokemonIndex]
}

func isAlive(pokemon *model.Pokemon) bool {
	return pokemon.Stats.HP > 0
}

func switchPokemon(pokemonsList []model.Pokemon, conn *net.UDPConn, addr *net.UDPAddr) *model.Pokemon {
	for i := 0; i < len(pokemonsList); i++ {
		if isAlive(&pokemonsList[i]) {
			return &pokemonsList[i]
		}
	}
	return nil
}

func displayCommandsList(conn *net.UDPConn, addr *net.UDPAddr) {
	conn.WriteToUDP([]byte("List of commands:\n"), addr)
	conn.WriteToUDP([]byte("\tattack: to attack the opponent\n"), addr)
	conn.WriteToUDP([]byte("\tswitch: to switch to another pokemon\n"), addr)
	conn.WriteToUDP([]byte("\tsurrender: to stop the game\n"), addr)
}

func displaySelectedPokemons(pokemonsList []model.Pokemon, conn *net.UDPConn, addr *net.UDPAddr) {
	conn.WriteToUDP([]byte("You have:\n"), addr)
	for i, pokemon := range pokemonsList {
		conn.WriteToUDP([]byte(fmt.Sprintf("%d. %s\n", i, pokemon.Name)), addr)
	}
	conn.WriteToUDP([]byte("Please enter the index of the pokemon you want to switch to:\n"), addr)
}

func switchToChosenPokemon(pokemonsList []model.Pokemon, conn *net.UDPConn, addr *net.UDPAddr) *model.Pokemon {
	for {
		index := readIndex(conn, addr)
		if index < 0 || index >= len(pokemonsList) {
			conn.WriteToUDP([]byte("Please enter a valid index.\n"), addr)
			continue
		}
		if isAlive(&pokemonsList[index]) {
			return &pokemonsList[index]
		} else {
			conn.WriteToUDP([]byte("This pokemon is dead. Please select another one.\n"), addr)
		}
	}
}

func readCommands(conn *net.UDPConn, addr *net.UDPAddr) string {
	buffer := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buffer)
	command := strings.TrimSpace(string(buffer[:n]))
	if command == "attack" || command == "switch" || command == "surrender" || command == "?" {
		return strings.ToLower(command)
	}
	conn.WriteToUDP([]byte("Please enter a valid command\n"), addr)
	return readCommands(conn, addr)
}

func readIndex(conn *net.UDPConn, addr *net.UDPAddr) int {
	buffer := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buffer)
	input := strings.TrimSpace(string(buffer[:n]))
	index, _ := strconv.Atoi(input)
	return index
}

func PrintPokemonInfo(index int, pokemon model.Pokemon) {
	fmt.Println(index, ":", pokemon.Name)

	fmt.Println("\tElements: ")
	for _, element := range pokemon.Elements {
		fmt.Println("\t\tElement:", element)
	}
	fmt.Println("\tEV:", pokemon.EV)
	fmt.Println("\tEvolutionLevel:", pokemon.EvolutionLevel)
	fmt.Println("\tNextEvolution:", pokemon.NextEvolution)

	fmt.Println("\tStats:")
	fmt.Println("\t\tHP:", pokemon.Stats.HP)
	fmt.Println("\t\tAttack:", pokemon.Stats.Attack)
	fmt.Println("\t\tDefense:", pokemon.Stats.Defense)
	fmt.Println("\t\tSpeed:", pokemon.Stats.Speed)
	fmt.Println("\t\tSp_Attack:", pokemon.Stats.Sp_Attack)
	fmt.Println("\t\tSp_Defense:", pokemon.Stats.Sp_Defense)

	fmt.Println("\tDamage When Attacked:")
	for _, element := range pokemon.DamegeWhenAttacked {
		fmt.Printf("\t\tElement: %s. Coefficient: %f\n", element.Element, element.Coefficient)
	}
}

func selectPokemon(player *model.Player, conn *net.UDPConn, addr *net.UDPAddr) *[]model.Pokemon {
	var selectedPokemons = []model.Pokemon{}
	msg := player.Name + " please select 3 pokemons from:\n"
	fmt.Printf(msg)
	_, err := conn.WriteToUDP([]byte(msg), addr)
	if err != nil {
		return nil
	}
	for i := range player.Inventory {
		PrintPokemonInfo(i, player.Inventory[i])
		_, err := conn.WriteToUDP([]byte(fmt.Sprintf("%d: %s\n", i, player.Inventory[i].Name)), addr)
		if err != nil {
			return nil
		}
	}

	counter := 1
	for {
		if len(selectedPokemons) == 3 {
			break
		}
		conn.WriteToUDP([]byte(fmt.Sprintf("Enter the index of the %d pokemon you want to select: ", counter)), addr)
		index := readIndex(conn, addr)
		if index < 0 || index >= len(player.Inventory) {
			conn.WriteToUDP([]byte("Invalid index\n"), addr)
			continue
		}

		if isContain(selectedPokemons, player.Inventory[index]) {
			conn.WriteToUDP([]byte("You have selected this pokemon. Please select another one.\n"), addr)
			continue
		}

		conn.WriteToUDP([]byte(fmt.Sprintf("Selected %s\n", player.Inventory[index].Name)), addr)
		counter++
		selectedPokemons = append(selectedPokemons, player.Inventory[index])
	}

	conn.WriteToUDP([]byte("You have selected: "), addr)
	for _, pokemon := range selectedPokemons {
		conn.WriteToUDP([]byte(fmt.Sprintf("%s ", pokemon.Name)), addr)
	}
	conn.WriteToUDP([]byte("\n"), addr)

	return &selectedPokemons
}

func RandomInt(max int64) (int64, error) {
	// Generate a random big integer in the range [0, max)
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err // Return the error if any
	}
	return n.Int64(), nil // Convert the big integer to int64 and return
}
