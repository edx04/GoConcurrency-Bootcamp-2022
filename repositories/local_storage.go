package repositories

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"GoConcurrency-Bootcamp-2022/models"
)

type LocalStorage struct{}

type result struct {
	Pokemon models.Pokemon
	Error   error
}

const filePath = "resources/pokemons.csv"

func (l LocalStorage) Write(pokemons []models.Pokemon) error {
	file, fErr := os.Create(filePath)
	if fErr != nil {
		return fErr
	}
	defer file.Close()

	w := csv.NewWriter(file)
	records := buildRecords(pokemons)
	if err := w.WriteAll(records); err != nil {
		return err
	}

	return nil
}

func (l LocalStorage) Read() ([]models.Pokemon, error) {
	file, fErr := os.Open(filePath)

	if fErr != nil {
		return nil, fErr
	}

	defer file.Close()

	done := make(chan interface{})
	defer close(done)

	results := parseCSVData(done, genRows(done, file))

	pokemons := []models.Pokemon{}
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		pokemons = append(pokemons, result.Pokemon)
	}

	return pokemons, nil
}

func genRows(done <-chan interface{}, f io.Reader) chan []string {
	out := make(chan []string)
	go func() {
		r := csv.NewReader(f)
		_, err := r.Read()
		if err != nil {
			log.Fatal(err)

		}
		for {
			row, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
				continue

			}

			select {
			case <-done:
				return

			case out <- row:
			}

		}
		close(out)

	}()

	return out
}

func buildRecords(pokemons []models.Pokemon) [][]string {
	headers := []string{"id", "name", "height", "weight", "flat_abilities"}
	records := [][]string{headers}
	for _, p := range pokemons {
		record := fmt.Sprintf("%d,%s,%d,%d,%s",
			p.ID,
			p.Name,
			p.Height,
			p.Weight,
			p.FlatAbilityURLs)
		records = append(records, strings.Split(record, ","))
	}

	return records
}

func parseCSVData(done <-chan interface{}, records <-chan []string) <-chan result {
	out := make(chan result)

	go func() {
		defer close(out)
		for r := range records {
			var err_ error
			err_ = nil
			id, err := strconv.Atoi(r[0])
			if err != nil {
				err_ = err
			}
			height, err := strconv.Atoi(r[2])
			if err != nil {
				err_ = err
			}
			weight, err := strconv.Atoi(r[3])
			if err != nil {
				err_ = err
			}
			pokemon := models.Pokemon{
				ID:              id,
				Name:            r[1],
				Height:          height,
				Weight:          weight,
				Abilities:       nil,
				FlatAbilityURLs: r[4],
				EffectEntries:   nil,
			}

			select {
			case <-done:
				return

			case out <- result{Pokemon: pokemon, Error: err_}:
			}

		}

	}()

	return out
}
