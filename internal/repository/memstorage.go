package repository

import (
	"fmt"
	"strconv"

	"github.com/qesterrx/tplmetrics/internal/model"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	mm := MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}

	return &mm
}

func (ms *MemStorage) UpdateMetric(metrica *model.Metrica) error {

	//А не будет ли это слабым местом, все таки при каждом сохранении будем два раза искать в мапе?
	savedMetrica, err := ms.GetMetric(metrica.Name)

	if err == nil && savedMetrica.Kind != metrica.Kind {
		return fmt.Errorf("metric %s have already registred as %s", metrica.Name, savedMetrica.Kind)
	}

	if metrica.Kind == model.Gauge {

		valStr := fmt.Sprintf("%v", metrica.Value)
		val, err := strconv.ParseFloat(valStr, 64)

		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("type of value for Gauge metric is wrong (expected float64), got value %s", metrica.Value)
		}

		ms.gauge[metrica.Name] = val
		return nil

	} else if metrica.Kind == model.Counter {

		valStr := fmt.Sprintf("%v", metrica.Value)
		val, err := strconv.ParseInt(valStr, 10, 64)

		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("type of value for Counter metric is wrong (expected int64), got value %s", metrica.Value)
		}

		ms.counter[metrica.Name] = ms.counter[metrica.Name] + val
		return nil

	} else {
		return fmt.Errorf("unimplemented type %s", metrica.Kind)
	}
}

func (ms *MemStorage) GetMetric(name string) (*model.Metrica, error) {

	valueG, ok := ms.gauge[name]

	if ok {
		return &model.Metrica{
			Name:  name,
			Kind:  model.Gauge,
			Value: valueG,
		}, nil
	}

	valueC, ok := ms.counter[name]

	if ok {
		return &model.Metrica{
			Name:  name,
			Kind:  model.Counter,
			Value: valueC,
		}, nil
	}

	return nil, fmt.Errorf("metric %s not found", name)

}

func (ms *MemStorage) ShowAllMetric() {
	fmt.Println("-------------GAUGE-------------")
	for k, v := range ms.gauge {
		fmt.Printf("%s : %5.5f \n", k, v)
	}
	fmt.Println("-------------COUNTER-------------")
	for k, v := range ms.counter {
		fmt.Printf("%s : %d \n", k, v)
	}
}

func (ms *MemStorage) GetAllMetric() *[]model.Metrica {

	var mm []model.Metrica

	for k, v := range ms.gauge {

		mm = append(mm, model.Metrica{
			Name:  k,
			Kind:  model.Gauge,
			Value: v,
		})
	}

	for k, v := range ms.counter {

		mm = append(mm, model.Metrica{
			Name:  k,
			Kind:  model.Counter,
			Value: v,
		})
	}

	return &mm
}
