# PokeGo

A simple, text-based version of Pokemon game: PokeCat and PokeBat. There are two modules of the game Cat -> Catching and Bat -> Battling. Written in [Go](https://golang.org) using [Ebitengine](https://ebiten.org) for the PokeCat's UI and UDP Protocol for communication.

## Requirements

- Go 1.22.0

## Development

All commands must be run from the `../PokeGo/PokeCat` directory.

Start the server, go to `server` folder:

```
go run .\server.go
```

Start the client, go to `client` folder:

```
go run .
```

## Project Structure
### Pokedex
Refer to the pokemon database website: [Pokedex](https://pokedex.org/) to build up the pokemon database for our games. The pokedex is built using [Playwright](https://github.com/playwright-community/playwright-go) to crawl the database from the mentioned website. The pokemon database is stored in `../Pokedex/pokedex.json`.

### PokeBat
- The PokeBat is a turn based battle, allows two persons to attend via network. Each player will pick 3 pokemons from their pokemon list (inventory) to join the battle.
- Each pokemon of winning player will get 1/3 total of accumulated exp of all pokemons on the losing team.

### PokeCat
- Pokeworld is a 100x100 cells created using [Tiled](https://www.mapeditor.org/) and [Ebitengine](https://ebiten.org).
- Player starts at a random coordinate. Player can move up/left/down/right one cell using WASD from keyboard or automatically moving if auto mode is enabled.
- Server spawns 50 pokemons each wave from the pokedex, spawned pokemons have random level and random EV point (0.5-1). Pokemon will be despawned after 5 mins w/o being captured.

### Others
| Folder       | Go Package | Description                                                                                                                 |                                         
| --------     | ---------- | --------------------------------------------------------------------------------------------------------------------------- |
| /animations  | ✅         | Contains the animations of the sprite                                                                                       |
| /assets      | ✅         | Contains built assets that are embedded into the final executable. This includes sprite, images, and maps files.            |
| /constants   | ✅         | Contains the constants using during the development of PokeCat and Pokeworld                                                |
| /entities    | ✅         | Contains the struct of the sprites using in the Pokeworld (Player and Pokeballs)                                            |
| /models      | ✅         | Contains the structure of the player and pokedex json data file                                                             |
| /spritesheet | ✅         | Exports the tile from the `Tileset.png`                                                                                     |
