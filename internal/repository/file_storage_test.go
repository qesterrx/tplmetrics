package repository

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_UpdateMetrica(t *testing.T) {
	// Создаем временный файл для тестов
	tmpFile, err := os.CreateTemp("", "file_storage_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeSync, false)
	require.NoError(t, err)

	// --- Base Counter
	err = fs.UpdateMetrica(model.NewMetricaCounter("c1", 1))
	require.NoError(t, err)

	err = fs.UpdateMetrica(model.NewMetricaCounter("c1", 1))
	assert.NoError(t, err)

	c1, err := fs.GetMetrica("c1", string(model.Counter))
	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaCounter(int64(2)), c1.Value())

	// --- Base Gauge
	err = fs.UpdateMetrica(model.NewMetricaGauge("g1", 1.0001))
	require.NoError(t, err)

	err = fs.UpdateMetrica(model.NewMetricaGauge("g1", 2.2222))
	assert.NoError(t, err)

	g1, err := fs.GetMetrica("g1", string(model.Gauge))
	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaGauge(float64(2.2222)), g1.Value())

	// Проверяем, что данные записались в файл
	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var metrics []model.MetricaJSONAdapter
	err = json.Unmarshal(data, &metrics)
	require.NoError(t, err)
	assert.Len(t, metrics, 2)
}

func TestFileStorage_UpdateMetrica_AsyncMode(t *testing.T) {
	// Тест для асинхронного режима (без автоматической записи в файл)
	tmpFile, err := os.CreateTemp("", "file_storage_async_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeAsync, false)
	require.NoError(t, err)

	err = fs.UpdateMetrica(model.NewMetricaCounter("c1", 10))
	require.NoError(t, err)

	// В асинхронном режиме данные не должны автоматически записаться в файл
	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Empty(t, data) // Файл должен быть пустым

	// Ручная синхронизация
	err = fs.WriteMetrics()
	require.NoError(t, err)

	// Теперь данные должны быть в файле
	data, err = os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestFileStorage_GetMetric(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "file_storage_get_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	counterValue := int64(5)
	gaugeValue := float64(2.2222)

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeSync, false)
	require.NoError(t, err)

	err = fs.UpdateMetrica(model.NewMetricaCounter("c1", counterValue))
	require.NoError(t, err)
	err = fs.UpdateMetrica(model.NewMetricaGauge("g1", gaugeValue))
	require.NoError(t, err)

	mtrk, err := fs.GetMetrica("c1", string(model.Counter))
	assert.NoError(t, err)
	assert.Equal(t, model.Counter, mtrk.Kind())
	assert.Equal(t, model.FormatMetricaCounter(counterValue), mtrk.Value())

	mtrk, err = fs.GetMetrica("g1", string(model.Gauge))
	assert.NoError(t, err)
	assert.Equal(t, model.Gauge, mtrk.Kind())
	assert.Equal(t, model.FormatMetricaGauge(gaugeValue), mtrk.Value())

	_, err = fs.GetMetrica("notfound", "counter")
	assert.Error(t, err)
}

func TestFileStorage_RestoreFromFile(t *testing.T) {
	// Сначала создаем файл с данными
	tmpFile, err := os.CreateTemp("", "file_storage_restore_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Записываем тестовые данные в файл
	testMetrics := []model.Metrica{
		model.NewMetricaCounter("c_restore", 100),
		model.NewMetricaGauge("g_restore", 99.99),
	}

	data, err := json.Marshal(testMetrics)
	require.NoError(t, err)

	err = os.WriteFile(tmpFile.Name(), data, 0666)
	require.NoError(t, err)
	tmpFile.Close()

	// Создаем новый FileStorage с restore=true
	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeAsync, true)
	require.NoError(t, err)

	// Проверяем, что данные восстановились
	c1, err := fs.GetMetrica("c_restore", string(model.Counter))
	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaCounter(100), c1.Value())

	g1, err := fs.GetMetrica("g_restore", string(model.Gauge))
	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaGauge(99.99), g1.Value())
}

func TestFileStorage_RestoreFromEmptyFile(t *testing.T) {
	// Создаем пустой файл
	tmpFile, err := os.CreateTemp("", "file_storage_empty_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Создаем FileStorage с restore=true и пустым файлом
	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeAsync, true)
	require.NoError(t, err)

	// Проверяем, что хранилище пустое
	metrics := fs.GetAllMetrics()
	assert.Empty(t, metrics)
}

func TestFileStorage_UpdateMetricaBatch(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "file_storage_batch_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeSync, false)
	require.NoError(t, err)

	batch := []model.Metrica{
		model.NewMetricaCounter("batch_c1", 10),
		model.NewMetricaCounter("batch_c1", 20), // суммарно должно стать 30
		model.NewMetricaGauge("batch_g1", 1.5),
		model.NewMetricaGauge("batch_g1", 2.5), // должно стать 2.5
	}

	err = fs.UpdateMetricaBatch(batch)
	require.NoError(t, err)

	c1, err := fs.GetMetrica("batch_c1", string(model.Counter))
	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaCounter(30), c1.Value())

	g1, err := fs.GetMetrica("batch_g1", string(model.Gauge))
	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaGauge(2.5), g1.Value())
}

func TestFileStorage_GetAllMetrics(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "file_storage_all_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeAsync, false)
	require.NoError(t, err)

	metrics := []model.Metrica{
		model.NewMetricaCounter("z_counter", 100),
		model.NewMetricaGauge("a_gauge", 99.99),
		model.NewMetricaCounter("m_counter", 50),
	}

	for _, m := range metrics {
		err = fs.UpdateMetrica(m)
		require.NoError(t, err)
	}

	allMetrics := fs.GetAllMetrics()
	assert.Len(t, allMetrics, 3)

	// Проверяем сортировку по имени (MemStorage.GetAllMetrics сортирует ключи)
	names := make([]string, len(allMetrics))
	for i, m := range allMetrics {
		names[i] = m.Name()
	}
	assert.Equal(t, []string{"m_counter", "z_counter", "a_gauge"}, names)
}

func TestFileStorage_WriteMetrics_OnlyWhenChanged(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "file_storage_changed_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeAsync, false)
	require.NoError(t, err)

	// Изначально файл пуст
	data, _ := os.ReadFile(tmpFile.Name())
	assert.Empty(t, data)

	// Вызов WriteMetrics без изменений - файл должен остаться пустым
	err = fs.WriteMetrics()
	require.NoError(t, err)
	data, _ = os.ReadFile(tmpFile.Name())
	assert.Empty(t, data)

	// Добавляем метрику
	err = fs.UpdateMetrica(model.NewMetricaCounter("c1", 42))
	require.NoError(t, err)

	// Теперь WriteMetrics должен записать данные
	err = fs.WriteMetrics()
	require.NoError(t, err)
	data, _ = os.ReadFile(tmpFile.Name())
	assert.NotEmpty(t, data)

	// Повторный вызов WriteMetrics без новых изменений не должен перезаписывать файл?
	// По логике hasChanged сброшен, поэтому повторная запись не произойдет
	modTime1, _ := os.Stat(tmpFile.Name())
	err = fs.WriteMetrics()
	require.NoError(t, err)
	modTime2, _ := os.Stat(tmpFile.Name())
	assert.Equal(t, modTime1.ModTime(), modTime2.ModTime())
}

func TestFileStorage_Check(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "file_storage_check_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ms := NewMemStorage()
	fs, err := NewFileStorage(ms, tmpFile.Name(), config.MetricaStorageModeAsync, false)
	require.NoError(t, err)

	err = fs.Check()
	assert.NoError(t, err)
}
