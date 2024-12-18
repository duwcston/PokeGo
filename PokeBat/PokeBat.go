package PokeBat

import (
	"PokeGo/model"
	"fmt"
	"net"
	"time"
)

func Battle(player1, player2 *model.Player, conn *net.UDPConn, addr1, addr2 *net.UDPAddr) (string, []model.Pokemon) {

	if player1 == nil {
		fmt.Println("Error: player1 is nil")
		return "", nil
	}
	if conn == nil {
		fmt.Println("Error: conn is nil")
		return "", nil
	}
	if addr1 == nil {
		fmt.Println("Error: addr1 is nil")
		return "", nil
	}
	if len(player1.Inventory) < 3 {
		fmt.Println("Player 1 has less than 3 pokemons")
		_, err := conn.WriteToUDP([]byte("You have less than 3 pokemons"), addr1)
		if err != nil {
			return "", nil
		}
		return "", nil
	} else if len(player2.Inventory) < 3 {
		fmt.Println("Player 2 has less than 3 pokemons")
		_, err := conn.WriteToUDP([]byte("You have less than 3 pokemons"), addr2)
		if err != nil {
			return "", nil
		}
		return "", nil
	}

	// Player 1 select 3 Pokemons
	player1Pokemons := selectPokemon(player1, conn, addr1)

	// Player 2 select 3 Pokemons
	player2Pokemons := selectPokemon(player2, conn, addr2)

	allBattlingPokemons := append(*player1Pokemons, *player2Pokemons...)
	firstAttacker := getHighestSpeedPokemon(allBattlingPokemons)
	var firstDefender *model.Pokemon

	fmt.Println("Battle start!")
	_, err := conn.WriteToUDP([]byte("Battle start!\n"), addr1)
	if err != nil {
		return "", nil
	}
	_, err = conn.WriteToUDP([]byte("Battle start!\n"), addr2)
	if err != nil {
		return "", nil
	}

	var IsTurnPlayer1 bool
	var IsTurnPlayer2 bool
	if isContain(*player1Pokemons, *firstAttacker) {
		firstDefender = getHighestSpeedPokemon(*player2Pokemons)
		fmt.Printf("%s goes first \n", player1.Name)
		conn.WriteToUDP([]byte(fmt.Sprintf("%s goes first \n", player1.Name)), addr1)
		conn.WriteToUDP([]byte(fmt.Sprintf("%s goes first \n", player1.Name)), addr2)
		IsTurnPlayer1 = true
		IsTurnPlayer2 = false
	} else {
		firstDefender = getHighestSpeedPokemon(*player1Pokemons)
		fmt.Printf("%s goes first \n", player2.Name)
		conn.WriteToUDP([]byte(fmt.Sprintf("%s goes first \n", player2.Name)), addr1)
		conn.WriteToUDP([]byte(fmt.Sprintf("%s goes first \n", player2.Name)), addr2)
		IsTurnPlayer1 = false
		IsTurnPlayer2 = true
	}
	var winnerName string
	var winPokemons []model.Pokemon
	var BattleEnd bool
	// The battle loop
	var player1Pokemon = firstAttacker
	var player2Pokemon = firstDefender
	for {
		if IsTurnPlayer1 {
			player1Pokemon, BattleEnd = PlayerMove(player1Pokemon, player2Pokemon, player1Pokemons, player1.Name, conn, addr1, addr2)
			if BattleEnd {
				winnerName = player2.Name
				winPokemons = LevelUpPokemon(player2.Inventory, *player2Pokemons)
				fmt.Println("end battel")
				break
			}
			IsTurnPlayer1 = false
			IsTurnPlayer2 = true
		}

		if IsTurnPlayer2 {
			player2Pokemon, BattleEnd = PlayerMove(player2Pokemon, player1Pokemon, player2Pokemons, player2.Name, conn, addr2, addr1)
			if BattleEnd {
				winnerName = player1.Name
				winPokemons = LevelUpPokemon(player1.Inventory, *player1Pokemons)
				fmt.Println("end battel")
				break
			}
			IsTurnPlayer2 = false
			IsTurnPlayer1 = true
		}

		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("Before")
	for i := range winPokemons {
		PrintPokemonInfo(i, winPokemons[i])
	}

	return winnerName, winPokemons

}

func LevelUpPokemon(Pokemons []model.Pokemon, BattlePokemon []model.Pokemon) []model.Pokemon {
	var BattlePokemonName []string
	for _, pokemon := range BattlePokemon {
		BattlePokemonName = append(BattlePokemonName, pokemon.Name)
	}
	for i := range Pokemons {
		if isContain(BattlePokemonName, Pokemons[i].Name) {
			Pokemons[i].Level += 20
		}
	}
	return Pokemons
}
