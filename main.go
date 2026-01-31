package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
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
	callback    func() error
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

var (
	locationAreasNextURL = "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"
	locationAreasPrevURL string
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

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

		if err := cmd.callback(); err != nil {
			fmt.Printf("Error executing command %s: %v\n", cmd.name, err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	fmt.Print("help: Display this help message\n")
	fmt.Print("map: Display the next 20 location areas\n")
	fmt.Print("mapb: Display the previous 20 location areas\n")
	fmt.Print("exit: Exit the Pokedex\n")
	return nil
}

func commandMap() error {
	if locationAreasNextURL == "" {
		fmt.Println("No more locations")
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Get(locationAreasNextURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("pokeapi returned status %s", res.Status)
	}

	var payload locationAreaListResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return err
	}

	for _, location := range payload.Results {
		fmt.Println(location.Name)
	}

	if payload.Next != nil {
		locationAreasNextURL = *payload.Next
	} else {
		locationAreasNextURL = ""
	}
	if payload.Previous != nil {
		locationAreasPrevURL = *payload.Previous
	} else {
		locationAreasPrevURL = ""
	}

	return nil
}

func commandMapb() error {
	if locationAreasPrevURL == "" {
		fmt.Println("No previous locations")
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Get(locationAreasPrevURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("pokeapi returned status %s", res.Status)
	}

	var payload locationAreaListResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return err
	}

	for _, location := range payload.Results {
		fmt.Println(location.Name)
	}

	if payload.Next != nil {
		locationAreasNextURL = *payload.Next
	} else {
		locationAreasNextURL = ""
	}
	if payload.Previous != nil {
		locationAreasPrevURL = *payload.Previous
	} else {
		locationAreasPrevURL = ""
	}

	return nil
}
