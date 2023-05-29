package utils

import (
	"context"
	"reflect"
	"time"
)

func LimitRateReflect(f interface{}, interval time.Duration) func(...interface{}) []interface{} {
	// Use closures to save the time of the last function call
	var lastCall time.Time

	fValue := reflect.ValueOf(f)
	fType := fValue.Type()

	if fType.Kind() != reflect.Func {
		panic("f must be a function")
	}

	//if fType.NumOut() == 0 {
	//	panic("f must have at least one output parameter")
	//}

	outCount := fType.NumOut()
	outTypes := make([]reflect.Type, outCount)

	for i := 0; i < outCount; i++ {
		outTypes[i] = fType.Out(i)
	}

	// Returns a new function, which is used to limit the function to be called only once at a specified time interval
	return func(args ...interface{}) []interface{} {
		// Calculate the time interval since the last function call
		elapsed := time.Since(lastCall)
		// If the interval is less than the specified time, wait for the remaining time
		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}
		// Update the time of the last function call
		lastCall = time.Now()

		inCount := fType.NumIn()
		in := make([]reflect.Value, inCount)

		if len(args) != inCount {
			panic("wrong number of arguments")
		}

		for i := 0; i < inCount; i++ {
			in[i] = reflect.ValueOf(args[i])
		}

		out := fValue.Call(in)

		if len(out) != outCount {
			panic("function returned wrong number of values")
		}

		result := make([]interface{}, outCount)

		for i := 0; i < outCount; i++ {
			result[i] = out[i].Interface()
		}

		return result
	}
}

type Fn[T any, R any] func(T) (R, error)
type FnCtx[T any, R any] func(context.Context, T) (R, error)

func LimitRate[T any, R any](f Fn[T, R], interval time.Duration) Fn[T, R] {
	// Use closures to save the time of the last function call
	var lastCall time.Time
	// Returns a new function, which is used to limit the function to be called only once at a specified time interval
	return func(t T) (R, error) {
		// Calculate the time interval since the last function call
		elapsed := time.Since(lastCall)
		// If the interval is less than the specified time, wait for the remaining time
		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}
		// Update the time of the last function call
		lastCall = time.Now()
		// Execute the function that needs to be limited
		return f(t)
	}
}

func LimitRateCtx[T any, R any](f FnCtx[T, R], interval time.Duration) FnCtx[T, R] {
	// Use closures to save the time of the last function call
	var lastCall time.Time
	// Returns a new function, which is used to limit the function to be called only once at a specified time interval
	return func(ctx context.Context, t T) (R, error) {
		// Calculate the time interval since the last function call
		elapsed := time.Since(lastCall)
		// If the interval is less than the specified time, wait for the remaining time
		if elapsed < interval {
			t := time.NewTimer(interval - elapsed)
			select {
			case <-ctx.Done():
				t.Stop()
				var zero R
				return zero, ctx.Err()
			case <-t.C:

			}
		}
		// Update the time of the last function call
		lastCall = time.Now()
		// Execute the function that needs to be limited
		return f(ctx, t)
	}
}
