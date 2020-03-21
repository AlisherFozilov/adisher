package di

import (
	"errors"
	"log"
	"reflect"
	"sort"
)

type Component struct {
	value reflect.Value
}

type container struct {
	registered []Constructor
	wired      []Component
}

func NewContainer() *container {
	return &container{registered: make([]Constructor, 0), wired: make([]Component, 0)}
}

type Constructor struct {
	numInArgs    int
	reflectValue reflect.Value
}

func (c *container) Provide(constructors ...interface{}) {
	c.register(constructors...)
	c.wire()
	log.Print(len(c.registered))
	log.Print(len(c.wired))

}

func (c *container) Component(ptrToComponentPtr interface{}) {
	//errors.As()
	if ptrToComponentPtr == nil {
		panic(errors.New("di: ptrToComponentPtr can't be nil"))
	}
	reflectValue := reflect.ValueOf(ptrToComponentPtr).Elem()
	reflectType := reflectValue.Type()
	//log.Print(reflectValue)
	//log.Print(reflectType)
	for _, component := range c.wired {
		if component.value.Type() == reflectType {
			reflectValue.Set(component.value)
			return
		}
	}
	panic(errors.New("no such component"))
}

func (c *container) register(constructors ...interface{}) {
	for _, constructor := range constructors {
		reflectType := reflect.TypeOf(constructor)

		if reflectType.Kind() != reflect.Func {
			panic(errors.New("one of components is not a function"))
		}

		if reflectType.NumOut() != 1 {
			panic(errors.New("number of out args is not 1"))
		}

		numIn := reflectType.NumIn()
		reflectValue := reflect.ValueOf(constructor)

		c.registered = append(c.registered, Constructor{
			numInArgs:    numIn,
			reflectValue: reflectValue,
		})
	}
}

func (c *container) wire() {
	sort.Slice(c.registered, func(i, j int) bool {
		return c.registered[i].numInArgs < c.registered[j].numInArgs
	})

	unwiredConstructors := make([]Constructor, 0)
	for _, constructor := range c.registered {
		if constructor.numInArgs == 0 {
			createdComponent := constructor.reflectValue.Call(nil)[0]
			c.wired = append(c.wired, Component{value: createdComponent})
		} else {
			unwiredConstructors = append(unwiredConstructors, constructor)
		}
	}

	iterationsNumber := len(unwiredConstructors)
	stillUnwiredConstructor := make([]Constructor, 0)

	for i := 0; i < iterationsNumber; i++ {

		for _, constructor := range unwiredConstructors {

			args := make([]reflect.Value, 0)
			for j := 0; j < constructor.numInArgs; j++ {
				argType := constructor.reflectValue.Type().In(j)
				argValue, ok := c.searchTypeInValuesReflection(argType)
				if !ok {
					break
				}
				args = append(args, argValue)
			}

			if len(args) == constructor.numInArgs {
				component := constructor.reflectValue.Call(args)[0]
				c.wired = append(c.wired, Component{value: component})
			} else {
				stillUnwiredConstructor = append(stillUnwiredConstructor, constructor)
			}
		}

		unwiredConstructors = stillUnwiredConstructor
		stillUnwiredConstructor = make([]Constructor, 0)
	}

	if len(unwiredConstructors) != 0 {
		log.Print("no dependencies for components below")
		for _, constructor := range unwiredConstructors {
			log.Print(constructor.reflectValue.Type())
		}
		panic("no dependencies for components above")
	}
}

func (c *container) searchTypeInValuesReflection(argType reflect.Type) (reflect.Value, bool) {

	for _, component := range c.wired {
		if component.value.Type() == argType {
			return component.value, true
		}
	}
	return reflect.Value{}, false
}
