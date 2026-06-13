package defval

import (
	"testing"
)

func BenchmarkCheckDefaultValue(b *testing.B) {

	b.Run("CheckDefaultValue (reflection)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// В math/rand/v2 нет встроенной функции заполнения слайса байт.
			_ = DVR("benchmark_string", "subs")
			_ = DVR("", "subs")
		}
	})

	b.Run("SimpleComparing", func(b *testing.B) {
		//Простейший аналог без рефлексии
		fn := func(a, b string) string {
			if a == "" {
				return b
			}
			return a
		}

		for i := 0; i < b.N; i++ {
			// В math/rand/v2 нет встроенной функции заполнения слайса байт.
			_ = fn("benchmark_string", "subs")
			_ = fn("", "subs")
		}
	})

	// BenchmarkCheckDefaultValue/CheckDefaultValue_(reflection)-16         	18193891	         65.23 ns/op	      16 B/op	       1 allocs/op
	// BenchmarkCheckDefaultValue/SimpleComparing-16                        	1000000000	         0.1128 ns/op	       0 B/op	       0 allocs/op
	//
	// Итог понятен, но для интереса оставлю вариант с рефлексией
}

func TestDVR(t *testing.T) {
	//Базовые типы с нулевыми значениями
	t.Run("basic types with zero values", func(t *testing.T) {
		// int
		if got := DVR(0, 42); got != 42 {
			t.Errorf("DVR(0, 42) = %v, want 42", got)
		}
		if got := DVR(10, 42); got != 10 {
			t.Errorf("DVR(10, 42) = %v, want 10", got)
		}

		// string
		if got := DVR("", "default"); got != "default" {
			t.Errorf(`DVR("", "default") = %v, want "default"`, got)
		}
		if got := DVR("hello", "default"); got != "hello" {
			t.Errorf(`DVR("hello", "default") = %v, want "hello"`, got)
		}

		// bool
		if got := DVR(false, true); got != true {
			t.Errorf("DVR(false, true) = %v, want true", got)
		}
		if got := DVR(true, false); got != true {
			t.Errorf("DVR(true, false) = %v, want true", got)
		}

		// float64
		if got := DVR(0.0, 3.14); got != 3.14 {
			t.Errorf("DVR(0.0, 3.14) = %v, want 3.14", got)
		}
		if got := DVR(2.71, 3.14); got != 2.71 {
			t.Errorf("DVR(2.71, 3.14) = %v, want 2.71", got)
		}
	})

	//Тут можно много всего еще проверять, но не нужно ))
}
