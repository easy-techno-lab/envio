package envio

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
)

const getError = "get data into"

func (e *engine) get(v any) (err error) {
	if t := reflect.ValueOf(v).Kind(); t != reflect.Pointer {
		return fmt.Errorf("%s: the input value is not a pointer", name)
	}

	s := e.newGetState()
	defer getStatePool.Put(s)

	s.get(v)
	return s.err
}

type getterState struct {
	*engine
	context
	*bytes.Buffer
}

var getStatePool sync.Pool

func (e *engine) newGetState() *getterState {
	if p := getStatePool.Get(); p != nil {
		s := p.(*getterState)
		s.err = nil
		s.Reset()
		return s
	}

	s := &getterState{engine: e, Buffer: new(bytes.Buffer)}
	s.field = new(field)
	return s
}

func (s *getterState) get(v any) {
	if err := s.reflectValue(reflect.ValueOf(v)); err != nil {
		if !errors.Is(err, errExist) {
			s.setError(name, getError, err)
		}
	}
}

func (s *getterState) reflectValue(v reflect.Value) error {
	s.context.field.typ = v.Type()
	return s.cachedFunctions(s.context.field.typ).getterFunc(s, v)
}

func (s *getterState) getEnv() error {
	str := os.Getenv(s.field.name)
	if s.field.mandatory && str == "" {
		s.err = fmt.Errorf("%s: the required variable $%s is missing", name, s.field.name)
		return errExist
	}
	s.WriteString(str)
	return nil
}

type getterFunc func(*getterState, reflect.Value) error

func (f *structFields) get(s *getterState, v reflect.Value) (err error) {
	s.structName = v.Type().Name()

	for _, s.field = range *f {
		s.Reset()
		rv := v.Field(s.field.index)

		if s.field.embedded != nil {
			if rv.Kind() == reflect.Pointer {
				if rv.IsNil() {
					s.err = fmt.Errorf("%s: %w: %s", name, ErrPointerToUnexported, rv.Type().Elem())
					return errExist
				}
				rv = rv.Elem()
			}

			if err = s.field.embedded.get(s, rv); err != nil {
				return
			}
			continue
		}

		if err = s.field.functions.getterFunc(s, rv); err != nil {
			return
		}
	}

	return
}

func boolGetter(s *getterState, v reflect.Value) error {
	if err := s.getEnv(); err != nil {
		return err
	}
	if s.Len() == 0 {
		return nil
	}
	r, err := strconv.ParseBool(s.String())
	v.SetBool(r)
	return err
}

func intGetter(s *getterState, v reflect.Value) error {
	if err := s.getEnv(); err != nil {
		return err
	}
	if s.Len() == 0 {
		return nil
	}
	r, err := strconv.ParseInt(s.String(), 10, bitSize(v.Kind()))
	v.SetInt(r)
	return err
}

func uintGetter(s *getterState, v reflect.Value) error {
	if err := s.getEnv(); err != nil {
		return err
	}
	if s.Len() == 0 {
		return nil
	}
	r, err := strconv.ParseUint(s.String(), 10, bitSize(v.Kind()))
	v.SetUint(r)
	return err
}

func floatGetter(s *getterState, v reflect.Value) error {
	if err := s.getEnv(); err != nil {
		return err
	}
	if s.Len() == 0 {
		return nil
	}
	r, err := strconv.ParseFloat(s.String(), bitSize(v.Kind()))
	v.SetFloat(r)
	return err
}

//func arrayGetter(s *getterState, v reflect.Value) error {
//	return nil
//}

func interfaceGetter(s *getterState, v reflect.Value) error {
	if v.IsNil() {
		s.err = ErrNilInterface
		return errExist
	}
	return s.reflectValue(v.Elem())
}

//func mapGetter(s *getterState, v reflect.Value) error {
//	return nil
//}

func pointerGetter(s *getterState, v reflect.Value) error {
	if v.IsNil() {
		rv := reflect.New(v.Type().Elem())
		if err := s.reflectValue(rv.Elem()); err != nil {
			return err
		}
		if !isEmptyValue(rv.Elem()) {
			v.Set(rv)
		}
		return nil
	}
	return s.reflectValue(v.Elem())
}

func bytesGetter(s *getterState, v reflect.Value) error {
	if err := s.getEnv(); err != nil {
		return err
	}
	if s.Len() == 0 {
		return nil
	}
	v.SetBytes(s.Bytes())
	return nil
}

//func sliceGetter(s *getterState, v reflect.Value) error {
//	return nil
//}

func stringGetter(s *getterState, v reflect.Value) error {
	if err := s.getEnv(); err != nil {
		return err
	}
	if s.Len() == 0 {
		return nil
	}
	v.SetString(s.String())
	return nil
}

func structGetter(s *getterState, v reflect.Value) error {
	f := s.cachedFields(v.Type())
	return f.get(s, v)
}

func unsupportedTypeGetter(s *getterState, _ reflect.Value) error {
	s.err = ErrNotSupportType
	return errExist
}
