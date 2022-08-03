package use_cases

import (
	"context"
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type api interface {
	FetchPokemon(id int) (models.Pokemon, error)
}

type writer interface {
	Write(pokemons []models.Pokemon) error
}

type Fetcher struct {
	api     api
	storage writer
}

type result struct {
	Error   error
	Pokemon models.Pokemon
}

func NewFetcher(api api, storage writer) Fetcher {
	return Fetcher{api, storage}
}

func (f Fetcher) Fetch(from, to int) error {
	var pokemons []models.Pokemon
	ctx, cancel := context.WithCancel(context.Background())

	ch := f.generator(ctx, from, to)
	for result := range ch {
		if result.Error != nil {
			cancel()
			return result.Error
		}
		pokemons = append(pokemons, result.Pokemon)
	}

	return f.storage.Write(pokemons)
}

func (f Fetcher) generator(ctx context.Context, from, to int) <-chan result {
	ch := make(chan result)
	wg := sync.WaitGroup{}
	for id := from; id <= to; id++ {
		wg.Add(1)

		go func(ctx context.Context, i int) {
			defer wg.Done()
			pokemon, err := f.api.FetchPokemon(i)
			if err == nil {
				var flatAbilities []string
				for _, t := range pokemon.Abilities {
					flatAbilities = append(flatAbilities, t.Ability.URL)
				}
				pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")

			}

			res := result{Error: err, Pokemon: pokemon}

			select {
			case <-ctx.Done():
				return
			case ch <- res:
			}

		}(ctx, id)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch

}
