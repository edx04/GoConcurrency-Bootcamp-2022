package use_cases

import (
	"context"
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type reader interface {
	Read() ([]models.Pokemon, error)
}

type saver interface {
	Save(context.Context, []models.Pokemon) error
}

type fetcher interface {
	FetchAbility(string) (models.Ability, error)
}

type Refresher struct {
	reader
	saver
	fetcher
}

func NewRefresher(reader reader, saver saver, fetcher fetcher) Refresher {
	return Refresher{reader, saver, fetcher}
}

func (r Refresher) Refresh(ctx context.Context) error {

	out := r.generator(ctx)

	//Fan Out
	abilityOut := r.ability(out)
	abilityOut2 := r.ability(out)
	abilityOut3 := r.ability(out)
	pokemons := []models.Pokemon{}

	//Fan In
	for p := range fanIn(abilityOut, abilityOut2, abilityOut3) {
		pokemons = append(pokemons, p)
	}

	if err := r.Save(ctx, pokemons); err != nil {
		return err
	}

	return nil
}

func (r Refresher) generator(ctx context.Context) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	go func() {
		defer close(out)
		pokemons, err := r.Read()
		if err == nil {
			for _, pokemon := range pokemons {
				out <- pokemon
			}
		}

	}()

	return out

}

func (r Refresher) ability(pokemons <-chan models.Pokemon) <-chan models.Pokemon {
	out := make(chan models.Pokemon)

	go func() {
		for p := range pokemons {
			urls := strings.Split(p.FlatAbilityURLs, "|")
			var abilities []string
			for _, url := range urls {
				ability, err := r.FetchAbility(url)
				if err != nil {
					panic(err)
				}

				for _, ee := range ability.EffectEntries {
					abilities = append(abilities, ee.Effect)
				}
			}
			p.EffectEntries = abilities

			out <- p

		}
		close(out)
	}()

	return out

}

func fanIn(inputs ...<-chan models.Pokemon) <-chan models.Pokemon {
	var wg sync.WaitGroup
	out := make(chan models.Pokemon)

	wg.Add(len(inputs))

	for _, in := range inputs {
		go func(ch <-chan models.Pokemon) {
			defer wg.Done()
			for value := range ch {
				out <- value
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
