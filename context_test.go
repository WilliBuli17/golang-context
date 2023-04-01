package golang_context

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
func TestContext(t *testing.T) {
	background := context.Background()
	fmt.Println("context background,", background)

	todo := context.TODO()
	fmt.Println("context todo,", todo)
}

// ---------------------------------------------------------------------------------------------------------------------
func TestContextWirhValue(t *testing.T) {
	contextA := context.Background()

	contextB := context.WithValue(contextA, "b", "B")
	contextC := context.WithValue(contextA, "c", "C")

	contextD := context.WithValue(contextB, "d", "D")
	contextE := context.WithValue(contextB, "e", "E")

	contextF := context.WithValue(contextC, "f", "F")

	contextG := context.WithValue(contextF, "g", "G")

	fmt.Println(contextA)
	fmt.Println(contextB)
	fmt.Println(contextC)
	fmt.Println(contextD)
	fmt.Println(contextE)
	fmt.Println(contextF)
	fmt.Println(contextG)

	fmt.Println(contextF.Value("f")) // dapat - karena valuenya sendiri
	fmt.Println(contextF.Value("c")) // dapat valuenya parent - karena value parentnya dia
	fmt.Println(contextF.Value("b")) // tidak dapat - karena key b tidak di jangkau oleh context f
	fmt.Println(contextA.Value("b")) // tidak dapat - karena aturan dalam kontext parent tidak bisa mengambil value childnya -- value itu selalu nanya ke atas, gak biisa nanya ke bawah
}

// ---------------------------------------------------------------------------------------------------------------------
func CreateCounter() chan int { // contoh go routine tidak pernah berhenti
	destination := make(chan int)

	go func() {
		defer close(destination)

		counter := 1
		for { // penyebabnya perulangan ini tidak pernah berhenti
			destination <- counter
			counter++
		}
	}()

	return destination
}

func TestGoroutineLeak(t *testing.T) { // contoh go routine tidak pernah berhenti atau tidak pernah mati
	fmt.Println(runtime.NumGoroutine())

	destination := CreateCounter()
	for n := range destination {
		fmt.Println("Counter,", n)
		if n == 10 {
			break
		}
	}

	fmt.Println(runtime.NumGoroutine())
}

// -- pemecahan masalah diatas

func CreateCounterx(ctx context.Context) chan int { // tambah parameter context
	destination := make(chan int)

	go func() {
		defer close(destination)

		counter := 1
		for {
			select { // pada tahap ini tambah pengeceka apakah udah done blm contextnya
			case <-ctx.Done():
				return
			default:
				destination <- counter
				counter++
			}
		}
	}()

	return destination
}

func TestContextWirhCancle(t *testing.T) {
	fmt.Println(runtime.NumGoroutine())

	parent := context.Background()
	ctx, cancel := context.WithCancel(parent)

	destination := CreateCounterx(ctx)
	fmt.Println(runtime.NumGoroutine())

	for n := range destination {
		fmt.Println("Counter,", n)
		if n == 10 {
			break
		}
	}

	cancel()
	time.Sleep(5 * time.Second) // cuma untuk memastika go routinenya udah mati -- tidak wajib
	fmt.Println(runtime.NumGoroutine())
}

// ---------------------------------------------------------------------------------------------------------------------
func CreateCountery(ctx context.Context) chan int { // tambah parameter context
	destination := make(chan int)

	go func() {
		defer close(destination)

		counter := 1
		for {
			select { // pada tahap ini tambah pengeceka apakah udah done blm contextnya
			case <-ctx.Done():
				return
			default:
				destination <- counter
				counter++
				time.Sleep(1 * time.Second) // slow simulation -- ini akan menyebabkan timeout / deadline
			}
		}
	}()

	return destination
}

func TestContextWirhTimeout(t *testing.T) { // tidak terlalu beda jauh dengan cancel, ada sedikit penambahan
	fmt.Println(runtime.NumGoroutine())

	parent := context.Background()
	ctx, cancel := context.WithTimeout(parent, 5*time.Second) // menentukan waktu timeout (cth 5 detik) -- saat menggunakan timeout, wajib juga menggunakan cancel
	defer cancel()                                            // harus menggunakan defer saat menggunakan timeout

	destination := CreateCountery(ctx)
	fmt.Println(runtime.NumGoroutine())

	for n := range destination {
		fmt.Println("Counter,", n)
	}

	time.Sleep(2 * time.Second) // cuma untuk memastika go routinenya udah mati -- tidak wajib
	fmt.Println(runtime.NumGoroutine())
}

// ---------------------------------------------------------------------------------------------------------------------
func TestContextWirhDeadline(t *testing.T) { // tidak terlalu beda jauh dengan timeout, ada sedikit penambahan
	fmt.Println(runtime.NumGoroutine())

	parent := context.Background()
	ctx, cancel := context.WithDeadline(parent, time.Now().Add(5*time.Second)) // menentukan waktu deadline (cth 5 detik dari saat di jalankan) -- saat menggunakan deadline, wajib juga menggunakan cancel
	defer cancel()                                                             // harus menggunakan defer saat menggunakan deadline

	destination := CreateCountery(ctx)
	fmt.Println(runtime.NumGoroutine())

	for n := range destination {
		fmt.Println("Counter,", n)
	}

	time.Sleep(2 * time.Second) // cuma untuk memastika go routinenya udah mati -- tidak wajib
	fmt.Println(runtime.NumGoroutine())
}
