package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipCompressMiddleware(t *testing.T) {

	gzip_func := func(msg []byte) ([]byte, error) {
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, err := gzw.Write([]byte(msg))
		if err != nil {
			return []byte{}, nil
		}
		err = gzw.Close()
		if err != nil {
			return []byte{}, nil
		}
		return buf.Bytes(), nil
	}

	//Получим ожидаемый результат
	request, err := gzip_func([]byte("Request"))
	assert.NoError(t, err)
	response, err := gzip_func([]byte("Request-Response"))
	assert.NoError(t, err)

	// Тестовый обработчик
	handler := GzipCompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res := string(req) + "-Response"
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(res))
	}))

	req := httptest.NewRequest("GET", "/", bytes.NewReader(request))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, response, body)

}
