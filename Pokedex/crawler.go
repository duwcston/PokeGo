package main

import (
	"PokeGo/model"
	"encoding/json"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	numberOfPokemons = 649
	baseURL          = "https://pokedex.org/#/"
)

var pokemons []model.Pokemon

func main() {
	crawlPokemonsDriver(numberOfPokemons)
}

func crawlPokemonsDriver(numsOfPokemons int) {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	page.Goto(baseURL)

	for i := range numsOfPokemons {

		locator := fmt.Sprintf("https://pokedex.org/#/pokemon/%d", i+1)
		page.Goto(locator)
		time.Sleep(500 * time.Millisecond)

		fmt.Print("Pokemon ", i+1, " ")
		pokemon := crawlPokemons(page)

		pokemons = append(pokemons, pokemon)

		page.Goto(baseURL)
		page.Reload()
	}

	// parse the pokemons variable to json file
	js, err := json.MarshalIndent(pokemons, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("pokedex.json", js, 0644)

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

func crawlStats(page playwright.Page) model.Stats {
	stats := model.Stats{}
	entries, _ := page.Locator("div.detail-panel-content > div.detail-header > div.detail-infobox > div.detail-stats > div.detail-stats-row").All()
	for _, entry := range entries {
		title, _ := entry.Locator("span:not([class])").TextContent()
		switch title {
		case "HP":
			hp, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.HP, _ = strconv.Atoi(hp)
		case "Attack":
			attack, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Attack, _ = strconv.Atoi(attack)
		case "Defense":
			defense, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Defense, _ = strconv.Atoi(defense)
		case "Speed":
			speed, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Speed, _ = strconv.Atoi(speed)
		case "Sp Atk":
			sp_Attack, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Sp_Attack, _ = strconv.Atoi(sp_Attack)
		case "Sp Def":
			sp_Defense, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Sp_Defense, _ = strconv.Atoi(sp_Defense)
		default:
			fmt.Println("Unknown title: ", title)
		}
	}
	return stats
}

func crawlProfile(page playwright.Page) model.Profile {
	genderRatio := model.GenderRatio{}
	profile := model.Profile{}
	entries, _ := page.Locator("div.detail-panel-content > div.detail-below-header > div.monster-minutia").All()
	for _, entry := range entries {
		title1, _ := entry.Locator("strong:not([class]):nth-child(1)").TextContent()
		stat1, _ := entry.Locator("span:not([class]):nth-child(2)").TextContent()
		switch title1 {
		case "Height:":
			heights := strings.Split(stat1, " ")
			height, _ := strconv.ParseFloat(heights[0], 32)
			profile.Height = float32(height)
		case "Catch Rate:":
			catchRates := strings.Split(stat1, "%")
			catchRate, _ := strconv.ParseFloat(catchRates[0], 32)
			profile.CatchRate = float32(catchRate)
		case "Egg Groups:":
			profile.EggGroup = stat1
		case "Abilities:":
			profile.Abilities = stat1
		}

		title2, _ := entry.Locator("strong:not([class]):nth-child(3)").TextContent()
		stat2, _ := entry.Locator("span:not([class]):nth-child(4)").TextContent()
		switch title2 {
		case "Weight:":
			weights := strings.Split(stat2, " ")
			weight, _ := strconv.ParseFloat(weights[0], 32)
			profile.Weight = float32(weight)
		case "Gender Ratio:":
			if stat2 == "N/A" {
				genderRatio.MaleRatio = 0
				genderRatio.FemaleRatio = 0
			} else {
				ratios := strings.Split(stat2, " ")

				maleRatios := strings.Split(ratios[0], "%")
				maleRatio, _ := strconv.ParseFloat(maleRatios[0], 32)
				genderRatio.MaleRatio = float32(maleRatio)

				femaleRatios := strings.Split(ratios[2], "%")
				femaleRatio, _ := strconv.ParseFloat(femaleRatios[0], 32)
				genderRatio.FemaleRatio = float32(femaleRatio)
			}

			profile.GenderRatio = genderRatio
		case "Hatch Steps:":
			profile.HatchSteps, _ = strconv.Atoi(stat2)
		}
	}
	return profile
}

func crawlDamegeWhenAttacked(page playwright.Page) []model.DamegeWhenAttacked {
	damegeWhenAttacked := []model.DamegeWhenAttacked{}
	entries, _ := page.Locator("div.when-attacked > div.when-attacked-row").All()
	for _, entry := range entries {
		element1, _ := entry.Locator("span.monster-type:nth-child(1)").TextContent()
		coefficient1, _ := entry.Locator("span.monster-multiplier:nth-child(2)").TextContent()
		coefficients1 := strings.Split(coefficient1, "x")
		coef1, _ := strconv.ParseFloat(coefficients1[0], 32)

		element2, _ := entry.Locator("span.monster-type:nth-child(3)").TextContent()
		coefficient2, _ := entry.Locator("span.monster-multiplier:nth-child(4)").TextContent()
		coefficients2 := strings.Split(coefficient2, "x")
		coef2, _ := strconv.ParseFloat(coefficients2[0], 32)

		damegeWhenAttacked = append(damegeWhenAttacked, model.DamegeWhenAttacked{Element: element1, Coefficient: float32(coef1)})
		damegeWhenAttacked = append(damegeWhenAttacked, model.DamegeWhenAttacked{Element: element2, Coefficient: float32(coef2)})
	}
	return damegeWhenAttacked
}

func crawlMoves(page playwright.Page) []model.Moves {
	moves := []model.Moves{}
	time.Sleep(500 * time.Millisecond)
	entries, _ := page.Locator("div.moves-row").All()
	if len(entries) != 0 {
		count := 0
		for _, entry := range entries {
			// simulate clicking the expand button in the move rows
			expandButton := page.Locator("div.moves-inner-row > button.dropdown-button").First()
			expandButton.Click()

			name, _ := entry.Locator("div.moves-inner-row > span:nth-child(2)").TextContent()
			element, _ := entry.Locator("div.moves-inner-row > span.monster-type").TextContent()

			powers, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(1)").TextContent()
			power := strings.Split(powers, ":")

			acc, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(2)").TextContent()
			accs := strings.Split(acc, ":")
			accValue := strings.Split(accs[1], "%")
			accInt, _ := strconv.Atoi(strings.ReplaceAll(accValue[0], " ", ""))

			pps, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(3)").TextContent()
			ppVal := strings.Split(pps, ":")
			pp, _ := strconv.Atoi(strings.ReplaceAll(ppVal[1], " ", ""))
			description := ""
			//description, _ := entry.Locator("div.moves-row-detail > div.move-description").TextContent()

			moves = append(moves, model.Moves{Name: name, Element: element, Power: power[1], Acc: accInt, PP: pp, Description: description})
			count++
			if count == 4 {
				return moves
			}
		}
	}
	return moves
}

func crawlEvolution(page playwright.Page, name string) (int, string) {
	entries, _ := page.Locator("div.evolutions > div.evolution-row").All()
	for _, entry := range entries {
		evolutionLabel, _ := entry.Locator("div.evolution-label > span").TextContent()
		evolutionLabels := strings.Split(evolutionLabel, " ")

		if evolutionLabels[0] == name {
			evolutionLevels := strings.Split(evolutionLabels[len(evolutionLabels)-1], ".")
			evolutionLevel, _ := strconv.Atoi(evolutionLevels[0])

			nextEvolution := evolutionLabels[3]
			return evolutionLevel, nextEvolution
		}
	}
	return 0, ""
}

func crawlElement(page playwright.Page) []string {
	elements := []string{}
	entries, _ := page.Locator("div.detail-types > span.monster-type").All()
	for _, entry := range entries {
		element, _ := entry.TextContent()
		elements = append(elements, element)
	}
	return elements
}

func crawlPokemons(page playwright.Page) model.Pokemon {
	pokemon := model.Pokemon{}
	name, _ := page.Locator("div.detail-panel > h1.detail-panel-header").TextContent()
	pokemon.Name = name
	pokemon.Stats = crawlStats(page)
	pokemon.Profile = crawlProfile(page)
	pokemon.DamegeWhenAttacked = crawlDamegeWhenAttacked(page)
	pokemon.EvolutionLevel, pokemon.NextEvolution = crawlEvolution(page, name)
	pokemon.Moves = crawlMoves(page)
	pokemon.Elements = crawlElement(page)
	fmt.Println(name, ": ", pokemon.Moves)
	return pokemon
}
