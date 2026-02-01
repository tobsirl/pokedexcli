package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
}

type config struct {
	Next     string
	Previous string
}

var commands = map[string]cliCommand{
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
	cfg := &config{Next: "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"}

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
	fmt.Print("map: Display the next 20 location areas\n")
	fmt.Print("mapb: Display the previous 20 location areas\n")
	fmt.Print("exit: Exit the Pokedex\n")
	return nil
}


func commandMap(cfg *config, _ ...string) error {
	if cfg.Next == "" {
		fmt.Println("No more locations")
		return nil
	}

	res, err := http.Get(cfg.Next)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("pokeapi returned status %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
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

	res, err := http.Get(cfg.Previous)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("pokeapi returned status %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
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
