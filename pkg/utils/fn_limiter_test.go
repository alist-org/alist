package utils_test

import (
	"context"
	"testing"
	"time"

	"github.com/alist-org/alist/v3/pkg/utils"
)

func myFunction(a int) (int, error) {
	// do something
	return a + 1, nil
}

func TestLimitRate(t *testing.T) {
	myLimitedFunction := utils.LimitRate(myFunction, time.Second)
	result, _ := myLimitedFunction(1)
	t.Log(result) // Output: 2
	result, _ = myLimitedFunction(2)
	t.Log(result) // Output: 3
}

type Test struct {
	limitFn func(string) (string, error)
}

func (t *Test) myFunction(a string) (string, error) {
	// do something
	return a + " world", nil
}

func TestLimitRateStruct(t *testing.T) {
	test := &Test{}
	test.limitFn = utils.LimitRate(test.myFunction, time.Second)
	result, _ := test.limitFn("hello")
	t.Log(result) // Output: hello world
	result, _ = test.limitFn("hi")
	t.Log(result) // Output: hi world
}

func myFunctionCtx(ctx context.Context, a int) (int, error) {
	// do something
	return a + 1, nil
}
func TestLimitRateCtx(t *testing.T) {
	myLimitedFunction := utils.LimitRateCtx(myFunctionCtx, time.Second)
	result, _ := myLimitedFunction(context.Background(), 1)
	t.Log(result) // Output: 2
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()
	result, err := myLimitedFunction(ctx, 2)
	t.Log(result, err) // Output: 0 context canceled
	result, _ = myLimitedFunction(context.Background(), 3)
	t.Log(result) // Output: 4
}
