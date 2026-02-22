package agent

import "sync"

//Вот это прям не знаю как выразиться... Для того чтобы не отправлять кучу собранных метрик типа "Значение" захотелось уйти от канала
//Вроде как логично использовать мапу, почему нет, но не тут то было
//Вся некрасота этого решения к метриках типа дельта. Получается что после отправки, если она была успешна то значение в мапе надо обнулить
//А вот если не успешна то обнулять не надо, т.к. сервер не получил эту дельту
//И если ее сейчас обнулить а на следующей итерации передать накопившуюся новую дельту то это уже введение в заблуждение сервера - а может я просто нашел проблему которой нет
//Впрочем, я думаю, что это было ошибкой, очень жду комментариев как это можно сделать лучше

type SafeMap struct {
	mutex     sync.RWMutex
	keys      map[string]bool
	values    map[string]float64
	deltas    map[string]int64
	lockedKey string
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		keys:   make(map[string]bool),
		values: make(map[string]float64),
		deltas: make(map[string]int64),
	}
}

func (sm *SafeMap) SetValue(key string, value float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.values[key] = value
	sm.keys[key] = true
}

func (sm *SafeMap) AddDelta(key string, delta int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	value, ok := sm.deltas[key]
	sm.keys[key] = true
	if ok {
		sm.deltas[key] = value + delta
	} else {
		sm.deltas[key] = delta
	}
}

func (sm *SafeMap) GetKeys() []string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	keys := []string{}
	for k, _ := range sm.keys {
		keys = append(keys, k)
	}
	return keys
}

func (sm *SafeMap) GetValue(key string) (float64, bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	value, ok := sm.values[key]
	return value, ok
}

func (sm *SafeMap) StartGetDeltaWithLock(key string) (int64, bool) {
	sm.mutex.Lock()
	value, ok := sm.deltas[key]
	if ok {
		sm.lockedKey = key
	} else {
		sm.mutex.Unlock()
	}

	return value, ok
}

func (sm *SafeMap) EndGetDeltaWithLock(confirm bool) {
	if confirm {
		sm.deltas[sm.lockedKey] = 0
	}
	sm.mutex.Unlock()
}
