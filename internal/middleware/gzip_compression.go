package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------compressWriter
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		// Создаем новый writer при необходимости
		return gzip.NewWriter(nil)
	},
}

// GZIPCompressWriter - структура необходимая для переопределения поведения http.ResponseWriter
type GZIPCompressWriter struct {
	// в данном случае мы используем не встраивание а композицию, поэтому надо определить все методы интерфейса под который мы мимикрируем
	w    http.ResponseWriter
	zipw *gzip.Writer
	pool *sync.Pool
}

// Дополнительные процедуры типа конструктор/деструктор
func newGZIPCompressWriter(dst http.ResponseWriter) *GZIPCompressWriter {
	zipw := gzipWriterPool.Get().(*gzip.Writer)
	zipw.Reset(dst)
	return &GZIPCompressWriter{w: dst, zipw: zipw, pool: &gzipWriterPool}
}

func (cw *GZIPCompressWriter) Close() error {
	err := cw.zipw.Close()
	cw.pool.Put(cw.zipw)
	return err
}

// Реализация интерфейса http.ResponseWriter
func (cw *GZIPCompressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *GZIPCompressWriter) Write(p []byte) (int, error) {
	return cw.zipw.Write(p)
}

func (cw *GZIPCompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}
	cw.w.WriteHeader(statusCode)
}

// ---------------------------------------------------------------------compressReader

type GZIPCompressReader struct {
	// в данном случае мы используем не встраивание а композицию, поэтому надо определить все методы интерфейса под который мы мимикрируем
	r  io.ReadCloser //этот интерфейс реализует *http.request.body
	zr *gzip.Reader
}

//Дополнительные процедуры типа конструктор

func newGZIPCompressReader(src io.ReadCloser) (*GZIPCompressReader, error) {

	nr, err := gzip.NewReader(src)
	if err != nil {
		return nil, err
	}

	return &GZIPCompressReader{
		r:  src,
		zr: nr,
	}, nil
}

// Реализация интерфейса
func (cr *GZIPCompressReader) Read(p []byte) (n int, err error) {
	return cr.zr.Read(p)
}

func (cr *GZIPCompressReader) Close() error {
	err := cr.r.Close()
	if err != nil {
		return err
	}

	return cr.zr.Close()
}

//---------------------------------------------------------------------Middleware

// GzipCompressMiddleware - Middleware-фунция обеспечивающая архивацию/разархивацию тела http.
// При разархивации описается на наличие заголовка в запросе клиента "Content-Encoding":"gzip"
// При архивации опирается на наличие заголовка в запросе клиента "Accept-Encoding":"gzip"
func GzipCompressMiddleware(h http.Handler) http.Handler {
	funcCompress := func(w http.ResponseWriter, r *http.Request) {

		finw := w

		//Подменяем райтер
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newGZIPCompressWriter(w)
			finw = cw
			defer cw.Close()
		}

		//Подменяем ридер
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newGZIPCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(finw, r)
	}

	return http.HandlerFunc(funcCompress)
}
