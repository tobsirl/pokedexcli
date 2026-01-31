package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

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

	// create a infinite loop to read user input
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cmdName := cleanInput(input)
		cmd, ok := commands[cmdName[0]]
		if !ok {
			fmt.Printf("Unknown command: %s\n", input)
			continue
		} else {
			err := cmd.callback()
			if err != nil {
				fmt.Printf("Error executing command %s: %v\n", cmd.name, err)
			}
		}
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
		fmt.Print("exit: Exit the Pokedex\n")
		return nil
}

func commandMap() error {
	res, err := http.Get("https://pokeapi.co/api/v2/location-area")
	if err != nil {
		return err
	}
	locations := struct {
	err := json.Unmarshal(res, &)
}