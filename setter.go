package envio

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"sync"
)

const setError = "set data from"

func (e *engine) set(v any) (err error) {
	s := e.newSetState()
	defer setStatePool.Put(s)

	s.set(v)
	return s.err
}

type setterState struct {
	*engine
	context
	scratch [64]byte
}

var setStatePool sync.Pool

func (e *engine) newSetState() *setterState {
	if p := setStatePool.Get(); p != nil {
		s := p.(*setterState)
		s.err = nil
		return s
	}

	s := &setterState{engine: e}
	s.field = new(field)
	return s
}

func (s *setterState) set(v any) {
	if err := s.reflectValue(reflect.ValueOf(v)); err != nil {
		if !errors.Is(err, errExist) {
			s.setError(name, setError, err)
		}
	}
}

func (s *setterState) reflectValue(v reflect.Value) error {
	s.context.field.typ = v.Type()
	return s.cachedFunctions(s.context.field.typ).setterFunc(s, v)
}

func (s *setterState) setEnv(v []byte) error {
	return os.Setenv(s.field.name, string(v))
}

type setterFunc func(*setterState, reflect.Value) error

func valueFromPtr(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Pointer {
		return v
	}
	if v.IsNil() {
		v = reflect.New(v.Type().Elem())
	}
	return v.Elem()
}

func (f *structFields) set(s *setterState, v reflect.Value) (err error) {
	s.structName = v.Type().Name()

	for _, s.field = range *f {
		rv := v.Field(s.field.index)

		// If the environment variable is mandatory,
		// then to avoid overwriting the value, ignore the field if it is empty.
		if s.field.mandatory && isEmptyValue(rv) {
			continue
		}

		if s.field.embedded != nil {
			if err = s.field.embedded.set(s, valueFromPtr(rv)); err != nil {
				return
			}
			continue
		}

		if err = s.field.functions.setterFunc(s, rv); err != nil {
			return
		}
	}

	return
}

func boolSetter(s *setterState, v reflect.Value) error {
	return s.setEnv(strconv.AppendBool(s.scratch[:0], v.Bool()))
}

func intSetter(s *setterState, v reflect.Value) error {
	return s.setEnv(strconv.AppendInt(s.scratch[:0], v.Int(), 10))
}

func uintSetter(s *setterState, v reflect.Value) error {
	return s.setEnv(strconv.AppendUint(s.scratch[:0], v.Uint(), 10))
}

func floatSetter(s *setterState, v reflect.Value) error {
	return s.setEnv(strconv.AppendFloat(s.scratch[:0], v.Float(), 'g', -1, bitSize(v.Kind())))
}

//func arraySetter(s *setterState, v reflect.Value) error {
//	return nil
//}

func interfaceSetter(s *setterState, v reflect.Value) error {
	if v.IsNil() {
		s.err = ErrNilInterface
		return errExist
	}
	return s.reflectValue(v.Elem())
}

//func mapSetter(s *setterState, v reflect.Value) error {
//	return nil
//}

func pointerSetter(s *setterState, v reflect.Value) error {
	return s.reflectValue(valueFromPtr(v))
}

func bytesSetter(s *setterState, v reflect.Value) error {
	return s.setEnv(v.Bytes())
}

//func sliceSetter(s *setterState, v reflect.Value) error {
//	return nil
//}

func stringSetter(s *setterState, v reflect.Value) error {
	return s.setEnv(append(s.scratch[:0], v.String()...))
}

func structSetter(s *setterState, v reflect.Value) error {
	f := s.cachedFields(v.Type())
	return f.set(s, reflect.ValueOf(v.Interface()))
}

func unsupportedTypeSetter(s *setterState, _ reflect.Value) error {
	s.err = ErrNotSupportType
	return errExist
}
