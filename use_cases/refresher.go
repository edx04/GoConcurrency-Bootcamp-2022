package use_cases

import (
	"context"
	"strings"

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

	out := r.generator()
	abilityOut := r.ability(out)
	pokemons := []models.Pokemon{}
	for p := range abilityOut {
		pokemons = append(pokemons, p)
	}
	// i := 0
	// for p := range out {
	// 	urls := strings.Split(p.FlatAbilityURLs, "|")
	// 	var abilities []string
	// 	for _, url := range urls {
	// 		ability, err := r.FetchAbility(url)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		for _, ee := range ability.EffectEntries {
	// 			abilities = append(abilities, ee.Effect)
	// 		}
	// 	}

	// 	pokemons[i].EffectEntries = abilities
	// 	i++
	// }

	if err := r.Save(ctx, pokemons); err != nil {
		return err
	}

	return nil
}

func (r Refresher) generator() <-chan models.Pokemon {
	out := make(chan models.Pokemon)

	go func() {
		pokemons, err := r.Read()

		if err != nil {
			panic(err)
		}

		for _, pokemon := range pokemons {
			out <- pokemon
		}

		defer close(out)

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
