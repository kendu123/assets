package digimon

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// Traits describe a Digimon's adaptive qualities in a 0-100 range.
type Traits struct {
	Power      float64
	Agility    float64
	Resilience float64
	Curiosity  float64
}

// Environment captures the changing conditions that shape evolution.
type Environment struct {
	Heat   float64
	Threat float64
	Social float64
}

// World simulates shifting conditions without requiring manual input each cycle.
type World struct {
	RNG      *rand.Rand
	Heat     float64
	Threat   float64
	Social   float64
	Volatile float64
}

// Digimon represents a living creature that adapts without a fixed evolution tree.
type Digimon struct {
	Name       string
	Generation int
	Age        int
	Traits     Traits
	History    []string
	Phenotype  Phenotype
}

// EvolutionEngine mutates and selects the best-fitting form per environment.
type EvolutionEngine struct {
	RNG               *rand.Rand
	VariantsPerCycle  int
	MutationMagnitude float64
}

// Phenotype summarizes the outward form derived from traits and the environment.
type Phenotype struct {
	Element     string
	Frame       string
	Temperament string
}

// NewEvolutionEngine creates an engine with deterministic randomness.
func NewEvolutionEngine(seed int64) *EvolutionEngine {
	return &EvolutionEngine{
		RNG:               rand.New(rand.NewSource(seed)),
		VariantsPerCycle:  6,
		MutationMagnitude: 12,
	}
}

// NewWorld creates a world that drifts over time.
func NewWorld(seed int64) *World {
	rng := rand.New(rand.NewSource(seed))
	return &World{
		RNG:      rng,
		Heat:     0.4 + rng.Float64()*0.2,
		Threat:   0.4 + rng.Float64()*0.2,
		Social:   0.4 + rng.Float64()*0.2,
		Volatile: 0.25,
	}
}

// NewDigimon creates a starting Digimon with a name and initial traits.
func NewDigimon(name string, traits Traits) Digimon {
	clampedTraits := clampTraits(traits)
	phenotype := describePhenotype(clampedTraits, Environment{})
	return Digimon{
		Name:       name,
		Generation: 1,
		Age:        0,
		Traits:     clampedTraits,
		Phenotype:  phenotype,
		History: []string{fmt.Sprintf(
			"Born as %s with %s (%s)",
			name,
			traitsSummary(traits),
			phenotypeSummary(phenotype),
		)},
	}
}

// Evolve mutates traits, evaluates fitness, and returns the best adapted form.
func (e *EvolutionEngine) Evolve(current Digimon, env Environment) (Digimon, string) {
	scaledMagnitude := e.MutationMagnitude * (1 + env.Threat*0.3 + env.Heat*0.2)
	best := current
	bestFitness := fitness(current.Traits, env)
	cycle := current.Age + 1

	for i := 0; i < e.VariantsPerCycle; i++ {
		candidateTraits := mutateTraits(current.Traits, e.RNG, scaledMagnitude)
		candidateFitness := fitness(candidateTraits, env)
		if candidateFitness > bestFitness {
			bestFitness = candidateFitness
			phenotype := describePhenotype(candidateTraits, env)
			best = Digimon{
				Name:       evolveName(current.Name, candidateTraits, env),
				Generation: current.Generation + 1,
				Age:        cycle,
				Traits:     candidateTraits,
				Phenotype:  phenotype,
				History:    append(append([]string{}, current.History...), ""),
			}
		}
	}

	log := fmt.Sprintf(
		"Cycle %d -> %s adapts (%s, %s) under heat %.2f / threat %.2f / social %.2f",
		cycle,
		best.Name,
		traitsSummary(best.Traits),
		phenotypeSummary(best.Phenotype),
		env.Heat,
		env.Threat,
		env.Social,
	)

	if best.Name == current.Name && bestFitness == fitness(current.Traits, env) {
		log = fmt.Sprintf(
			"Cycle %d -> %s holds form (%s, %s) under heat %.2f / threat %.2f / social %.2f",
			cycle,
			current.Name,
			traitsSummary(current.Traits),
			phenotypeSummary(current.Phenotype),
			env.Heat,
			env.Threat,
			env.Social,
		)
		current.Age = cycle
		current.History = append(current.History, log)
		return current, log
	}

	best.History[len(best.History)-1] = log
	return best, log
}

// NextEnvironment advances the world and returns the new environment snapshot.
func (w *World) NextEnvironment() Environment {
	shift := func(value float64) float64 {
		drift := (w.RNG.Float64() - 0.5) * 2 * w.Volatile
		return math.Max(0, math.Min(1, value+drift))
	}

	w.Heat = shift(w.Heat)
	w.Threat = shift(w.Threat)
	w.Social = shift(w.Social)

	return Environment{
		Heat:   w.Heat,
		Threat: w.Threat,
		Social: w.Social,
	}
}

// Simulate lets a Digimon adapt over multiple autonomous world cycles.
func (e *EvolutionEngine) Simulate(current Digimon, world *World, cycles int) Digimon {
	for i := 0; i < cycles; i++ {
		env := world.NextEnvironment()
		next, _ := e.Evolve(current, env)
		current = next
	}
	return current
}

func mutateTraits(traits Traits, rng *rand.Rand, magnitude float64) Traits {
	jitter := func(value float64) float64 {
		return value + rng.NormFloat64()*magnitude
	}

	return clampTraits(Traits{
		Power:      jitter(traits.Power),
		Agility:    jitter(traits.Agility),
		Resilience: jitter(traits.Resilience),
		Curiosity:  jitter(traits.Curiosity),
	})
}

func fitness(traits Traits, env Environment) float64 {
	heatWeight := env.Heat
	threatWeight := env.Threat
	socialWeight := env.Social

	return (traits.Resilience*0.6+traits.Power*0.4)*threatWeight +
		(traits.Resilience*0.7+traits.Agility*0.3)*heatWeight +
		(traits.Curiosity*0.6+traits.Agility*0.4)*socialWeight +
		(traits.Power+traits.Agility+traits.Resilience+traits.Curiosity)*0.05
}

func evolveName(current string, traits Traits, env Environment) string {
	dominant := dominantTrait(traits)
	prefix := traitPrefix(dominant)
	suffix := environmentSuffix(env)

	base := strings.TrimSpace(prefix + " " + current)
	return strings.TrimSpace(base + suffix)
}

func describePhenotype(traits Traits, env Environment) Phenotype {
	element := "Aether"
	if env.Heat > 0.6 {
		element = "Flare"
	} else if env.Threat > 0.6 {
		element = "Bastion"
	} else if env.Social > 0.6 {
		element = "Harmony"
	}

	frame := "Balanced"
	if traits.Power > 70 || traits.Resilience > 70 {
		frame = "Brute"
	} else if traits.Agility > 70 {
		frame = "Skirmisher"
	} else if traits.Curiosity > 70 {
		frame = "Mystic"
	}

	temperament := "Steady"
	if traits.Curiosity > 65 && env.Social > 0.5 {
		temperament = "Inquisitive"
	} else if traits.Power > 65 && env.Threat > 0.5 {
		temperament = "Fierce"
	} else if traits.Resilience > 65 {
		temperament = "Guarded"
	}

	return Phenotype{
		Element:     element,
		Frame:       frame,
		Temperament: temperament,
	}
}

func dominantTrait(traits Traits) string {
	maxTrait := traits.Power
	name := "power"
	if traits.Agility > maxTrait {
		maxTrait = traits.Agility
		name = "agility"
	}
	if traits.Resilience > maxTrait {
		maxTrait = traits.Resilience
		name = "resilience"
	}
	if traits.Curiosity > maxTrait {
		name = "curiosity"
	}
	return name
}

func traitPrefix(trait string) string {
	switch trait {
	case "power":
		return "Volt"
	case "agility":
		return "Swift"
	case "resilience":
		return "Stone"
	case "curiosity":
		return "Echo"
	default:
		return "Neo"
	}
}

func environmentSuffix(env Environment) string {
	if env.Threat > env.Heat && env.Threat > env.Social {
		return "-Aegis"
	}
	if env.Heat > env.Social {
		return "-Pyre"
	}
	if env.Social > 0.6 {
		return "-Link"
	}
	return ""
}

func clampTraits(traits Traits) Traits {
	clamp := func(value float64) float64 {
		return math.Max(0, math.Min(100, value))
	}

	return Traits{
		Power:      clamp(traits.Power),
		Agility:    clamp(traits.Agility),
		Resilience: clamp(traits.Resilience),
		Curiosity:  clamp(traits.Curiosity),
	}
}

func traitsSummary(traits Traits) string {
	return fmt.Sprintf(
		"P%.0f A%.0f R%.0f C%.0f",
		traits.Power,
		traits.Agility,
		traits.Resilience,
		traits.Curiosity,
	)
}

func phenotypeSummary(phenotype Phenotype) string {
	return fmt.Sprintf(
		"%s %s, %s",
		phenotype.Element,
		phenotype.Frame,
		phenotype.Temperament,
	)
}
