package genetics

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"sync"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/strategies"
)

const (
	PopulationSize = 100
	Generations    = 50
	MutationRate   = 0.2
)

type Score struct {
	Value      float64
	Individual strategies.StrategyWeights
}

func GenerateRandomWeights() strategies.StrategyWeights {
	return strategies.StrategyWeights{
		SuperTrendWeight:  rand.Float64() * 3,       // Range [0, 3]
		BollingerWeight:   rand.Float64() * 3,       // Range [0, 3]
		EmaWeight:         rand.Float64() * 3,       // Range [0, 3]
		RsiWeight:         rand.Float64() * 3,       // Range [0, 3]
		MacdWeight:        rand.Float64() * 3,       // Range [0, 3]
		AtrMultiplier:     rand.Float64()*2.5 + 1.5, // Range [1.5, 4]
		StrengthThreshold: rand.Float64() * 10,      // Range [0, 10]
	}
}

// Crossover combines two parents to create a child
func Crossover(parent1, parent2 strategies.StrategyWeights) strategies.StrategyWeights {
	return strategies.StrategyWeights{
		SuperTrendWeight:  (parent1.SuperTrendWeight + parent2.SuperTrendWeight) / 2,
		BollingerWeight:   (parent1.BollingerWeight + parent2.BollingerWeight) / 2,
		EmaWeight:         (parent1.EmaWeight + parent2.EmaWeight) / 2,
		RsiWeight:         (parent1.RsiWeight + parent2.RsiWeight) / 2,
		MacdWeight:        (parent1.MacdWeight + parent2.MacdWeight) / 2,
		StrengthThreshold: (parent1.StrengthThreshold + parent2.StrengthThreshold) / 2,
		AtrMultiplier:     (parent1.AtrMultiplier + parent2.AtrMultiplier) / 2,
	}
}

// Mutate applies random changes to a StrategyWeights
func Mutate(weights strategies.StrategyWeights) strategies.StrategyWeights {
	if rand.Float64() < MutationRate {
		weights.SuperTrendWeight += rand.Float64()*0.2 - 0.1
	}
	if rand.Float64() < MutationRate {
		weights.BollingerWeight += rand.Float64()*0.2 - 0.1
	}
	if rand.Float64() < MutationRate {
		weights.EmaWeight += rand.Float64()*0.2 - 0.1
	}
	if rand.Float64() < MutationRate {
		weights.RsiWeight += rand.Float64()*0.2 - 0.1
	}
	if rand.Float64() < MutationRate {
		weights.MacdWeight += rand.Float64()*0.2 - 0.1
	}
	if rand.Float64() < MutationRate {
		weights.StrengthThreshold += rand.Float64()*0.5 - 0.25
	}
	if rand.Float64() < MutationRate {
		weights.AtrMultiplier += rand.Float64()*0.4 - 0.2
	}

	if weights.SuperTrendWeight < 0 {
		weights.SuperTrendWeight = 0
	}
	if weights.BollingerWeight < 0 {
		weights.BollingerWeight = 0
	}
	if weights.EmaWeight < 0 {
		weights.EmaWeight = 0
	}
	if weights.RsiWeight < 0 {
		weights.RsiWeight = 0
	}
	if weights.MacdWeight < 0 {
		weights.MacdWeight = 0
	}
	if weights.StrengthThreshold < 0 {
		weights.StrengthThreshold = 0
	}
	if weights.AtrMultiplier < 1.5 {
		weights.AtrMultiplier = 1.5
	}

	if weights.SuperTrendWeight > 3 {
		weights.SuperTrendWeight = 3
	}
	if weights.BollingerWeight > 3 {
		weights.BollingerWeight = 3
	}
	if weights.EmaWeight > 3 {
		weights.EmaWeight = 3
	}
	if weights.RsiWeight > 3 {
		weights.RsiWeight = 3
	}
	if weights.MacdWeight > 3 {
		weights.MacdWeight = 3
	}
	if weights.StrengthThreshold > 10 {
		weights.StrengthThreshold = 10
	}
	if weights.AtrMultiplier > 4 {
		weights.AtrMultiplier = 4
	}

	return weights
}

func FitnessFunction(weights strategies.StrategyWeights, assets <-chan *asset.Snapshot) Score {
	var outcome float64
	scalp := strategies.Scalping{
		Weights:       weights,
		Stabilization: 100,
	}

	ac, oc := scalp.ComputeWithOutcome(assets, false)
	for o := range oc {
		<-ac
		outcome = o
	}

	// log.Printf("Outcome: %v", outcome)
	return Score{
		Value:      outcome,
		Individual: weights,
	}
}

func RunGenetic(repo asset.Repository, a string) error {
	log.Println("Running Genetic Algorithm")
	// Initialize population
	population := make([]strategies.StrategyWeights, PopulationSize)
	for i := range population {
		population[i] = GenerateRandomWeights()
	}

	// Genetic Algorithm
	for gen := 0; gen < Generations; gen++ {
		var wg sync.WaitGroup
		wg.Add(PopulationSize)
		fitnessScores := make([]Score, PopulationSize)

		log.Printf("Generation %d\n", gen)
		snapshots, err := repo.Get(a)
		if err != nil {
			return fmt.Errorf("Error getting BTC data: %v", err)
		}
		ss := helper.Duplicate(snapshots, PopulationSize)

		// Evaluate fitness
		for i, individual := range population {
			go func() {
				fitnessScores[i] = FitnessFunction(individual, ss[i])
				wg.Done()
			}()
		}

		wg.Wait()

		slices.SortFunc(fitnessScores, func(a, b Score) int {
			if a.Value < b.Value {
				return 1
			}
			if a.Value > b.Value {
				return -1
			}
			return 0
		})

		// Replace old population with new one
		population = generateNewPop(fitnessScores)
	}

	// // Final output
	// bestIndividual := SelectBest(population, fitnessScoresdcv )
	// fmt.Printf("Best Strategy Weights: %+v\n", bestIndividual)

	return nil
}

func generateNewPop(fitnessScores []Score) []strategies.StrategyWeights {
	newPopulation := make([]strategies.StrategyWeights, PopulationSize)

	// Elitism: Print top 5 individuals and add to next gen
	for i := 0; i < 5; i++ {
		fmt.Printf("Fitness: %.4f, Weights: %+v\n", fitnessScores[i].Value, fitnessScores[i].Individual)
		newPopulation[i] = fitnessScores[i].Individual
	}

	// Tournament selection for most
	for i := 5; i < PopulationSize-20; i++ {
		// Select 5 random individuals
		tournament := make([]Score, 5)
		for j := 0; j < 5; j++ {
			tournament[j] = fitnessScores[rand.IntN(PopulationSize)]
		}

		// Sort by fitness
		slices.SortFunc(tournament, func(a, b Score) int {
			if a.Value < b.Value {
				return 1
			}
			if a.Value > b.Value {
				return -1
			}
			return 0
		})

		parent1 := tournament[0].Individual
		parent2 := tournament[1].Individual

		// Crossover
		child := Crossover(parent1, parent2)

		// Mutate
		child = Mutate(child)

		newPopulation[i] = child
	}

	// last 20 individuals are random new individuals for diversity
	for i := PopulationSize - 20; i < PopulationSize; i++ {
		newPopulation[i] = GenerateRandomWeights()
	}

	return newPopulation
}
