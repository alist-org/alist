// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package singleflight

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group[string]
	v, err, _ := g.Do("key", func() (string, error) {
		return "bar", nil
	})
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v; want %v", got, want)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

func TestDoErr(t *testing.T) {
	var g Group[any]
	someErr := errors.New("Some error")
	v, err, _ := g.Do("key", func() (any, error) {
		return nil, someErr
	})
	if err != someErr {
		t.Errorf("Do error = %v; want someErr %v", err, someErr)
	}
	if v != nil {
		t.Errorf("unexpected non-nil value %#v", v)
	}
}

func TestDoDupSuppress(t *testing.T) {
	var g Group[string]
	var wg1, wg2 sync.WaitGroup
	c := make(chan string, 1)
	var calls int32
	fn := func() (string, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			// First invocation.
			wg1.Done()
		}
		v := <-c
		c <- v // pump; make available for any future calls

		time.Sleep(10 * time.Millisecond) // let more goroutines enter Do

		return v, nil
	}

	const n = 10
	wg1.Add(1)
	for i := 0; i < n; i++ {
		wg1.Add(1)
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			wg1.Done()
			v, err, _ := g.Do("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
				return
			}
			if v != "bar" {
				t.Errorf("Do = %T %v; want %q", v, v, "bar")
			}
		}()
	}
	wg1.Wait()
	// At least one goroutine is in fn now and all of them have at
	// least reached the line before the Do.
	c <- "bar"
	wg2.Wait()
	if got := atomic.LoadInt32(&calls); got <= 0 || got >= n {
		t.Errorf("number of calls = %d; want over 0 and less than %d", got, n)
	}
}

// Test that singleflight behaves correctly after Forget called.
// See https://github.com/golang/go/issues/31420
func TestForget(t *testing.T) {
	var g Group[any]

	var (
		firstStarted  = make(chan struct{})
		unblockFirst  = make(chan struct{})
		firstFinished = make(chan struct{})
	)

	go func() {
		g.Do("key", func() (i any, e error) {
			close(firstStarted)
			<-unblockFirst
			close(firstFinished)
			return
		})
	}()
	<-firstStarted
	g.Forget("key")

	unblockSecond := make(chan struct{})
	secondResult := g.DoChan("key", func() (i any, e error) {
		<-unblockSecond
		return 2, nil
	})

	close(unblockFirst)
	<-firstFinished

	thirdResult := g.DoChan("key", func() (i any, e error) {
		return 3, nil
	})

	close(unblockSecond)
	<-secondResult
	r := <-thirdResult
	if r.Val != 2 {
		t.Errorf("We should receive result produced by second call, expected: 2, got %d", r.Val)
	}
}

func TestDoChan(t *testing.T) {
	var g Group[string]
	ch := g.DoChan("key", func() (string, error) {
		return "bar", nil
	})

	res := <-ch
	v := res.Val
	err := res.Err
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v; want %v", got, want)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

// Test singleflight behaves correctly after Do panic.
// See https://github.com/golang/go/issues/41133
func TestPanicDo(t *testing.T) {
	var g Group[any]
	fn := func() (any, error) {
		panic("invalid memory address or nil pointer dereference")
	}

	const n = 5
	waited := int32(n)
	panicCount := int32(0)
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					t.Logf("Got panic: %v\n%s", err, debug.Stack())
					atomic.AddInt32(&panicCount, 1)
				}

				if atomic.AddInt32(&waited, -1) == 0 {
					close(done)
				}
			}()

			g.Do("key", fn)
		}()
	}

	select {
	case <-done:
		if panicCount != n {
			t.Errorf("Expect %d panic, but got %d", n, panicCount)
		}
	case <-time.After(time.Second):
		t.Fatalf("Do hangs")
	}
}

func TestGoexitDo(t *testing.T) {
	var g Group[any]
	fn := func() (any, error) {
		runtime.Goexit()
		return nil, nil
	}

	const n = 5
	waited := int32(n)
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			var err error
			defer func() {
				if err != nil {
					t.Errorf("Error should be nil, but got: %v", err)
				}
				if atomic.AddInt32(&waited, -1) == 0 {
					close(done)
				}
			}()
			_, err, _ = g.Do("key", fn)
		}()
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("Do hangs")
	}
}

func TestPanicDoChan(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skipf("js does not support exec")
	}

	if os.Getenv("TEST_PANIC_DOCHAN") != "" {
		defer func() {
			recover()
		}()

		g := new(Group[any])
		ch := g.DoChan("", func() (any, error) {
			panic("Panicking in DoChan")
		})
		<-ch
		t.Fatalf("DoChan unexpectedly returned")
	}

	t.Parallel()

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name(), "-test.v")
	cmd.Env = append(os.Environ(), "TEST_PANIC_DOCHAN=1")
	out := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	err := cmd.Wait()
	t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)
	if err == nil {
		t.Errorf("Test subprocess passed; want a crash due to panic in DoChan")
	}
	if bytes.Contains(out.Bytes(), []byte("DoChan unexpectedly")) {
		t.Errorf("Test subprocess failed with an unexpected failure mode.")
	}
	if !bytes.Contains(out.Bytes(), []byte("Panicking in DoChan")) {
		t.Errorf("Test subprocess failed, but the crash isn't caused by panicking in DoChan")
	}
}

func TestPanicDoSharedByDoChan(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skipf("js does not support exec")
	}

	if os.Getenv("TEST_PANIC_DOCHAN") != "" {
		blocked := make(chan struct{})
		unblock := make(chan struct{})

		g := new(Group[any])
		go func() {
			defer func() {
				recover()
			}()
			g.Do("", func() (any, error) {
				close(blocked)
				<-unblock
				panic("Panicking in Do")
			})
		}()

		<-blocked
		ch := g.DoChan("", func() (any, error) {
			panic("DoChan unexpectedly executed callback")
		})
		close(unblock)
		<-ch
		t.Fatalf("DoChan unexpectedly returned")
	}

	t.Parallel()

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name(), "-test.v")
	cmd.Env = append(os.Environ(), "TEST_PANIC_DOCHAN=1")
	out := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	err := cmd.Wait()
	t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)
	if err == nil {
		t.Errorf("Test subprocess passed; want a crash due to panic in Do shared by DoChan")
	}
	if bytes.Contains(out.Bytes(), []byte("DoChan unexpectedly")) {
		t.Errorf("Test subprocess failed with an unexpected failure mode.")
	}
	if !bytes.Contains(out.Bytes(), []byte("Panicking in Do")) {
		t.Errorf("Test subprocess failed, but the crash isn't caused by panicking in Do")
	}
}
