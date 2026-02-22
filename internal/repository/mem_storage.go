package repository

import (
	"fmt"
	"sort"

	"github.com/qesterrx/tplmetrics/internal/model"
)

type MemStorage struct {
	storage     map[string]model.Metrica
	keys        []string
	eventsQueue chan string
}

// Фабрика
func NewMemStorage() *MemStorage {
	mm := MemStorage{
		storage: make(map[string]model.Metrica),
		keys:    make([]string, 0),
	}

	return &mm
}

// Получение метрики по имени
func (ms *MemStorage) Metrica(name string, kind string) (model.Metrica, error) {
	key := kind + "_" + name
	mtrk, ok := ms.storage[key]

	if ok {
		return mtrk, nil
	}

	return nil, fmt.Errorf("metrica %s not found", key)

}

// Обновление метрики
func (ms *MemStorage) UpdateMetrica(mtrk model.Metrica) error {

	key := string(mtrk.Kind()) + "_" + mtrk.Name()

	mtrkSaved, ok := ms.storage[key]
	if !ok {
		ms.storage[key] = mtrk
		ms.keys = append(ms.keys, key)
		if ms.eventsQueue != nil {
			ms.eventsQueue <- "create"
		}
		return nil
	}

	err := mtrkSaved.UpdateValue(mtrk)
	if err != nil {
		return err
	}

	if ms.eventsQueue != nil {
		ms.eventsQueue <- "update"
	}

	return nil

}

// Получение всех сохраненных, с сортировкой по имени
func (ms *MemStorage) AllMetrics() []model.Metrica {

	sort.Strings(ms.keys)

	var mm []model.Metrica

	for _, v := range ms.keys {
		mm = append(mm, ms.storage[v])
	}

	return mm
}

// Показываем текущее состояние в output
func (ms *MemStorage) Debug() {

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

func (ms *MemStorage) GetQueueEvents() *chan string {
	ms.eventsQueue = make(chan string, 10000)
	return &ms.eventsQueue
}
