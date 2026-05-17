package repository

import (
	"fmt"
	"sort"
	"sync"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

// MemStorage - Базовая реализация интерфейса MetricaStorage
// MemStorage используется в других реализациях интерфейса MetricaStorage для реализации транзакицонности на уровне сервиса
type MemStorage struct {
	storage map[string]model.Metrica
	keys    []string
	mtx     sync.Mutex
}

// NewMemStorage - функция возвращает новый экземпляр MemStorage
func NewMemStorage() *MemStorage {
	logger.Log.Debug().Msg("Создание MemStorage")
	mm := MemStorage{
		storage: make(map[string]model.Metrica),
		keys:    make([]string, 0),
	}

	return &mm
}

// GetMetrica - возвращает метрику по имени и типу
func (ms *MemStorage) GetMetrica(name string, kind string) (model.Metrica, error) {

	//Так уж и быть поддержим одинаковые имена метрик разного типа
	key := kind + "_" + name
	mtrk, ok := ms.storage[key]

	if ok {
		return mtrk, nil
	}

	return nil, fmt.Errorf("metrica %s not found", key)

}

// UpdateMetrica - обновляет метрику
func (ms *MemStorage) UpdateMetrica(mtrk model.Metrica) error {

	key := string(mtrk.Kind()) + "_" + mtrk.Name()

	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	mtrkSaved, ok := ms.storage[key]
	if ok {
		err := mtrkSaved.UpdateValueAtomic(mtrk)
		if err != nil {
			return err
		}
	} else {
		ms.storage[key] = mtrk
		ms.keys = append(ms.keys, key)
	}

	return nil

}

// UpdateMetricaBatch Обновляет массив метрик
func (ms *MemStorage) UpdateMetricaBatch(mtrks []model.Metrica) error {

	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	touched, err := ms.startUpdateMetricaBatch(mtrks)
	if err != nil {
		ms.restoreUpdateMetricaBatch(touched)
		return err
	}

	ms.confirmUpdateMetricaBatch(touched)
	return nil
}

// AllMetrics - Получение всех сохраненных метрик с сортировкой по имени
func (ms *MemStorage) GetAllMetrics() []model.Metrica {

	sort.Strings(ms.keys)

	var mm []model.Metrica

	for _, v := range ms.keys {
		mm = append(mm, ms.storage[v])
	}

	return mm
}

// Debug - Показываем текущее состояние в output
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

// WriteMetrics - заглушка, при хранении в памяти писать некуда
func (ms *MemStorage) WriteMetrics() error {
	return nil
}

// Check - Проверка хранилища, заглушка, MemStorage всегда готов к работе
func (ms *MemStorage) Check() error {
	return nil
}

// т.к остальные реализации интерфейса MetricaStorage реализованы встраиванием этого
// нужны операции с возможностью отката и фиксации
// чет мне кажется что я упоролся.. но пока еще не понял до конца

// startUpdateMetricaBatch - Обновление массива метрик - а вот тут "массовая" операция, если наткнулись на ошибку - надо все откатить
func (ms *MemStorage) startUpdateMetricaBatch(mtrks []model.Metrica) (*[]*model.Metrica, error) {

	touched := []*model.Metrica{}

	for _, mtrk := range mtrks {

		key := string(mtrk.Kind()) + "_" + mtrk.Name()

		mtrkSaved, ok := ms.storage[key]
		if ok {
			err := mtrkSaved.UpdateValueStart(mtrk)
			if err != nil {
				return &touched, err
			}
			//Для возможности отката запоминаем объекты которые мы обновили
			touched = append(touched, &mtrkSaved)
		} else {
			ms.storage[key] = mtrk
			ms.keys = append(ms.keys, key)
			//Для возможности отката запоминаем объекты которые мы обновили
			touched = append(touched, &mtrk)
		}

	}

	return &touched, nil
}

// confirmUpdateMetricaBatch - Подтверждение массового изменения метрик
func (ms *MemStorage) confirmUpdateMetricaBatch(mtrks *[]*model.Metrica) {

	for _, mtrk := range *mtrks {
		(*mtrk).Confirm()
	}
}

// restoreUpdateMetricaBatch - Отмена массового изменения метрик
func (ms *MemStorage) restoreUpdateMetricaBatch(mtrks *[]*model.Metrica) {

	for _, mtrk := range *mtrks {
		(*mtrk).Restore()
	}
}
