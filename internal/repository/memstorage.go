package repository

import (
	"fmt"
	"sort"

	"github.com/qesterrx/tplmetrics/internal/model"
)

type MemStorage struct {
	storage map[string]model.Metrica
	keys    []string
}

func NewMemStorage() *MemStorage {
	mm := MemStorage{
		storage: make(map[string]model.Metrica),
		keys:    make([]string, 0),
	}

	return &mm
}

func (ms *MemStorage) Metrica(name string) (model.Metrica, error) {

	mtrk, ok := ms.storage[name]

	if ok {
		return mtrk, nil
	}

	return nil, fmt.Errorf("metric %s not found", name)

}

func (ms *MemStorage) UpdateMetrica(mtrk model.Metrica) error {

	name := mtrk.Name()
	mtrkSaved, ok := ms.storage[name]

	if !ok {
		ms.storage[name] = mtrk
		ms.keys = append(ms.keys, name)
		return nil
	}

	err := mtrkSaved.UpdateValue(mtrk)

	if err != nil {
		return err
	}

	return nil

}

func (ms *MemStorage) ShowDebug() {

	sort.Strings(ms.keys)

	fmt.Println("-------------KEYS-------------")
	for _, v := range ms.keys {
		fmt.Print(v, " ")
	}

	fmt.Println("-------------STORAGE-------------")
	for _, key := range ms.keys {
		mtrk := ms.storage[key]
		fmt.Printf("%s [%s]: %s \n", mtrk.Name(), mtrk.Kind(), mtrk.Value())
	}
}

func (ms *MemStorage) AllMetrica() []model.Metrica {

	sort.Strings(ms.keys)

	var mm []model.Metrica

	for _, v := range ms.keys {
		mm = append(mm, ms.storage[v])
	}

	return mm
}
