package compression

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

//---------------------------------------------------------------------compressWriter

// в данном случае мы используем не встраивание а композицию, поэтому надо определить все методы интерфейса под который мы мимикрируем
type GZIPCompressWriter struct {
	w    http.ResponseWriter
	zipw *gzip.Writer
}

// Дополнительные процедуры типа конструктор/деструктор
func newGZIPCompressWriter(dst http.ResponseWriter) *GZIPCompressWriter {
	return &GZIPCompressWriter{w: dst, zipw: gzip.NewWriter(dst)}
}

func (cw *GZIPCompressWriter) Close() error {
	return cw.zipw.Close()
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

//---------------------------------------------------------------------compressReader

// в данном случае мы используем не встраивание а композицию, поэтому надо определить все методы интерфейса под который мы мимикрируем
type GZIPCompressReader struct {
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
