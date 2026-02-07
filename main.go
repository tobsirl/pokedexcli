package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/tobsirl/pokedexcli/internal/pokecache"
)

type namedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type locationAreaListResponse struct {
	Count    int                `json:"count"`
	Next     *string            `json:"next"`
	Previous *string            `json:"previous"`
	Results  []namedAPIResource `json:"results"`
}

type pokemonEncounter struct {
	Pokemon namedAPIResource `json:"pokemon"`
}

type locationAreaResponse struct {
	Name              string             `json:"name"`
	PokemonEncounters []pokemonEncounter `json:"pokemon_encounters"`
}

type pokemonResponse struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

type Pokemon struct {
	Name           string
	BaseExperience int
	Height         int
	Weight         int
	Stats          map[string]int
	Types          []string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
}

type config struct {
	Next     string
	Previous string
	Cache    *pokecache.Cache
	Pokedex  map[string]Pokemon
	Rand     *rand.Rand
}

var commands = map[string]cliCommand{
	"catch": {
		name:        "catch",
		description: "Catch a Pokemon",
		callback:    commandCatch,
	},
	"inspect": {
		name:        "inspect",
		description: "Inspect a caught Pokemon",
		callback:    commandInspect,
	},
	"map": {
		name:        "map",
		description: "Display the Pokedex map",
		callback:    commandMap,
	},
	"mapb": {
		name:        "mapb",
		description: "Display the previous page of locations",
		callback:    commandMapb,
	},
	"explore": {
		name:        "explore",
		description: "Explore a location area",
		callback:    commandExplore,
	},
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex CLI",
		callback:    commandExit,
	},
	"help": {
		name:        "help",
		description: "Display this help message",
		callback:    commandHelp,
	},
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := &config{
		Next:  "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Cache: pokecache.NewCache(5 * time.Second),
		Pokedex: make(map[string]Pokemon),
		Rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// create an infinite loop to read user input
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		cmd, ok := commands[words[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := cmd.callback(cfg, words[1:]...); err != nil {
			fmt.Printf("Error executing command %s: %v\n", cmd.name, err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}

func commandExit(_ *config, _ ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *config, _ ...string) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	fmt.Print("help: Display this help message\n")
	fmt.Print("catch <pokemon_name>: Throw a Pokeball and try to catch a Pokemon\n")
	fmt.Print("inspect <pokemon_name>: View details about a caught Pokemon\n")
	fmt.Print("map: Display the next 20 location areas\n")
	fmt.Print("mapb: Display the previous 20 location areas\n")
	fmt.Print("explore <area_name>: Explore a location area\n")
	fmt.Print("exit: Exit the Pokedex\n")
	return nil
}

func catchChanceFromBaseExperience(baseExperience int) float64 {
	if baseExperience < 0 {
		baseExperience = 0
	}

	// Higher base experience => harder to catch.
	// Clamp into a reasonable range so common Pokemon are catchable within a few tries.
	chance := 1.0 - (float64(baseExperience) / 500.0)
	if chance < 0.05 {
		chance = 0.05
	}
	if chance > 0.85 {
		chance = 0.85
	}
	return chance
}

func commandCatch(cfg *config, args ...string) error {
	if len(args) < 1 {
		fmt.Println("usage: catch <pokemon_name>")
		return nil
	}

	pokemonName := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)

	body, err := getWithCache(cfg.Cache, url)
	if err != nil {
		return err
	}

	var payload pokemonResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	// Must print this before deciding caught vs escaped.
	fmt.Printf("Throwing a Pokeball at %s...\n", payload.Name)

	rng := cfg.Rand
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	chance := catchChanceFromBaseExperience(payload.BaseExperience)
	if rng.Float64() < chance {
		fmt.Printf("%s was caught!\n", payload.Name)
		if cfg.Pokedex == nil {
			cfg.Pokedex = make(map[string]Pokemon)
		}

		stats := make(map[string]int, len(payload.Stats))
		for _, s := range payload.Stats {
			stats[s.Stat.Name] = s.BaseStat
		}
		types := make([]string, 0, len(payload.Types))
		for _, t := range payload.Types {
			types = append(types, t.Type.Name)
		}

		cfg.Pokedex[payload.Name] = Pokemon{
			Name:           payload.Name,
			BaseExperience: payload.BaseExperience,
			Height:         payload.Height,
			Weight:         payload.Weight,
			Stats:          stats,
			Types:          types,
		}
		return nil
	}

	fmt.Printf("%s escaped!\n", payload.Name)
	return nil
}

func commandInspect(cfg *config, args ...string) error {
	if len(args) < 1 {
		fmt.Println("usage: inspect <pokemon_name>")
		return nil
	}

	pokemonName := args[0]
	if cfg == nil || cfg.Pokedex == nil {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	p, ok := cfg.Pokedex[pokemonName]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Height: %d\n", p.Height)
	fmt.Printf("Weight: %d\n", p.Weight)

	fmt.Println("Stats:")
	statOrder := []string{"hp", "attack", "defense", "special-attack", "special-defense", "speed"}
	for _, statName := range statOrder {
		if val, ok := p.Stats[statName]; ok {
			fmt.Printf("  -%s: %d\n", statName, val)
		}
	}

	fmt.Println("Types:")
	for _, t := range p.Types {
		fmt.Printf("  - %s\n", t)
	}

	return nil
}


func commandMap(cfg *config, _ ...string) error {
	if cfg.Next == "" {
		fmt.Println("No more locations")
		return nil
	}

	body, err := getWithCache(cfg.Cache, cfg.Next)
	if err != nil {
		return err
	}

	var payload locationAreaListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	for _, location := range payload.Results {
		fmt.Println(location.Name)
	}

	if payload.Next != nil {
		cfg.Next = *payload.Next
	} else {
		cfg.Next = ""
	}
	if payload.Previous != nil {
		cfg.Previous = *payload.Previous
	} else {
		cfg.Previous = ""
	}

	return nil
}


func commandMapb(cfg *config, _ ...string) error {
	if cfg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	body, err := getWithCache(cfg.Cache, cfg.Previous)
	if err != nil {
		return err
	}

	var payload locationAreaListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	for _, location := range payload.Results {
		fmt.Println(location.Name)
	}

	if payload.Next != nil {
		cfg.Next = *payload.Next
	} else {
		cfg.Next = ""
	}
	if payload.Previous != nil {
		cfg.Previous = *payload.Previous
	} else {
		cfg.Previous = ""
	}

	return nil
}

func commandExplore(cfg *config, args ...string) error {
	if len(args) < 1 {
		fmt.Println("usage: explore <area_name>")
		return nil
	}

	areaName := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)

	fmt.Printf("Exploring %s...\n", areaName)

	body, err := getWithCache(cfg.Cache, url)
	if err != nil {
		return err
	}

	var payload locationAreaResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range payload.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func getWithCache(cache *pokecache.Cache, url string) ([]byte, error) {
	if cache != nil {
		if val, ok := cache.Get(url); ok {
			return val, nil
		}
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("pokeapi returned status %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if cache != nil {
		cache.Add(url, body)
	}

	return body, nil
}
