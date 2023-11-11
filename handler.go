package fernet

import (
	"context"
	"fmt"
	"reflect"
)

func createHandler[T RequestContext](fn any) func(context.Context, T) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("handlers must be a function")
	}

	// if the parameters are `context.Context, T` then we can just call the function directly
	if goodFN, ok := fn.(func(context.Context, T)); ok {
		return goodFN
	}

	toPass := make([]func(context.Context, T) (bool, reflect.Value), 0, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		param := fnType.In(i)
		switch {
		case param.ConvertibleTo(reflect.TypeOf((*context.Context)(nil)).Elem()):
			toPass = append(toPass, func(ctx context.Context, rc T) (bool, reflect.Value) {
				return true, reflect.ValueOf(ctx)
			})
		case param.ConvertibleTo(reflect.TypeOf((*T)(nil)).Elem()):
			toPass = append(toPass, func(ctx context.Context, rc T) (bool, reflect.Value) {
				return true, reflect.ValueOf(rc)
			})
		case param.Implements(reflect.TypeOf((*FromRequest[T])(nil)).Elem()):
			toPass = append(toPass, func(ctx context.Context, rc T) (bool, reflect.Value) {
				realParamValue := reflect.New(param)
				if param.Kind() == reflect.Ptr {
					realParamValue = realParamValue.Elem()
				}
				realParam := realParamValue.Interface()

				isOK := realParam.(FromRequest[T]).FromRequest(ctx, rc)
				if !isOK {
					// does this always panic?
					return false, reflect.ValueOf(nil)
				}

				return true, reflect.ValueOf(realParam)
			})

		default:
			t := reflect.TypeOf((*T)(nil)).Elem()
			if param.Implements(reflect.TypeOf((*RequestContext)(nil)).Elem()) {
				panic(fmt.Sprintf(
					"received RequestContext type %s, but expected %s",
					param,
					reflect.TypeOf((*T)(nil)).Elem(),
				))
			}

			_, implementsFromRequest := param.MethodByName("FromRequest")
			if !implementsFromRequest && param.Kind() != reflect.Ptr {
				if _, ok := reflect.PtrTo(param).MethodByName("FromRequest"); ok {
					panic(
						fmt.Sprintf(
							"%s of %s, does not implement FromRequest[%s]. FromRequest has pointer receiver, but %s is not a pointer",
							param,
							fnType,
							t,
							param,
						),
					)
				}
			}

			if implementsFromRequest {
				panic(
					fmt.Sprintf(
						"FromRequest method on %s of %s, must have the signature `func(context.Context, %s) bool. Got `%s`",
						param,
						fnType,
						t,
						param,
					),
				)
			}

			panic(
				fmt.Sprintf(
					"paramter %d (%s) in function %s is not a valid type, must be context.Context, %s, or implement FromRequest[%s]",
					i+1,
					param,
					fnType,
					t,
					t,
				),
			)
		}
	}

	return func(ctx context.Context, req T) {
		params := make([]reflect.Value, len(toPass))
		paramsOK := true

		for i, fn := range toPass {
			ok, value := fn(ctx, req)
			if !ok {
				paramsOK = false
				break
			}

			params[i] = value
		}

		if paramsOK {
			reflect.ValueOf(fn).Call(params)
		}
	}
}
