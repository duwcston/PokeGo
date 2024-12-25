package PokeBat

import (
	"PokeGo/model"
	"fmt"
	"net"
)

func EvolutionProcess(winner *model.Player, Pokemon []model.Pokemon, AllPokemons []model.Pokemon, conn *net.UDPConn) {
	for i := range Pokemon {
		if Pokemon[i].Level >= Pokemon[i].EvolutionLevel && Pokemon[i].NextEvolution != "" {
			conn.WriteToUDP([]byte(fmt.Sprintf("%s evolves into %s\n", Pokemon[i].Name, Pokemon[i].NextEvolution)), winner.Addr)
			Pokemon[i] = Evolution(Pokemon[i], AllPokemons)
		}
	}
	winner.Inventory = Pokemon
	fmt.Println("After")
	for i := range Pokemon {
		PrintPokemonInfo(i, Pokemon[i])
	}
}
func Evolution(EVPokemon model.Pokemon, AllPokemons []model.Pokemon) model.Pokemon {
	for _, pokemon := range AllPokemons {
		if EVPokemon.NextEvolution == pokemon.Name {
			pokemon.Level = EVPokemon.Level
			return pokemon
		}
	}
	return EVPokemon
}
