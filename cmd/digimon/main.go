package main

import (
	"fmt"
	"time"

	"github.com/trustwallet/assets/internal/digimon"
)

func main() {
	engine := digimon.NewEvolutionEngine(time.Now().UnixNano())
	world := digimon.NewWorld(time.Now().UnixNano())
	starter := digimon.NewDigimon("Lumon", digimon.Traits{
		Power:      52,
		Agility:    58,
		Resilience: 46,
		Curiosity:  62,
	})

	current := engine.Simulate(starter, world, 6)

	fmt.Println("Evolution history:")
	for _, entry := range current.History {
		fmt.Printf("- %s\n", entry)
	}
}
