package main

import (
	"errors"
	"flag"
	"log"
	"reflect"
	"strings"
)

// GetTypesRecursive links in namesmap a flag with there flildstruct Type
// You can whether provide objValue on a structure or a pointer to structure as first argument
// Flags are genereted from field name or from structags
func GetTypesRecursive(objValue reflect.Value, namesmap map[string]reflect.Type, key string) error {
	name := key
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		for i := 0; i < objValue.NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("description"); len(tag) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
					if _, ok := namesmap[strings.ToLower(tag)]; ok {
						return errors.New("Tag already exists: " + tag)
					}
					namesmap[strings.ToLower(tag)] = objValue.Field(i).Type()
				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}
				if _, ok := namesmap[strings.ToLower(name)]; ok {
					return errors.New("Tag already exists: " + name)
				}
				namesmap[strings.ToLower(name)] = objValue.Field(i).Type()
				if err := GetTypesRecursive(objValue.Field(i), namesmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Ptr:
		typ := objValue.Type().Elem()
		inst := reflect.New(typ).Elem()
		if err := GetTypesRecursive(inst, namesmap, name); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

//ParseArgs : parses args into a map[tag]value, using map[type]parser
//args must be formated as like as flag documentation. See https://golang.org/pkg/flag
func ParseArgs(args []string, tagsmap map[string]reflect.Type, parsers map[reflect.Type]flag.Value) map[string]interface{} {
	newParsers := map[string]flag.Value{}
	flagSet := flag.NewFlagSet("flaeg.ParseArgs", flag.ExitOnError)
	valmap := make(map[string]interface{})
	for tag, rType := range tagsmap {
		if parser, ok := parsers[rType]; ok {
			newparser := reflect.New(reflect.TypeOf(parser).Elem()).Interface().(flag.Value)
			flagSet.Var(newparser, tag, "help")
			newParsers[tag] = newparser
		}
	}
	flagSet.Parse(args)
	for tag, newParser := range newParsers {
		valmap[tag] = newParser
	}
	return valmap
}

//FillStructRecursive initialize a value of any taged Struct given by reference
func FillStructRecursive(objValue reflect.Value, valmap map[string]interface{}, key string) error {
	name := key
	// fmt.Printf("objValue begin : %+v\n", objValue)
	switch objValue.Kind() {
	case reflect.Struct:
		name += objValue.Type().Name()
		inst := reflect.New(objValue.Type()).Elem()
		for i := 0; i < inst.NumField(); i++ {
			if tag := objValue.Type().Field(i).Tag.Get("description"); len(tag) > 0 {
				fieldName := objValue.Type().Field(i).Name
				if tag := objValue.Type().Field(i).Tag.Get("long"); len(tag) > 0 {
					fieldName = tag
				}
				if tag := objValue.Type().Field(i).Tag.Get("short"); len(tag) > 0 {
					SetFields(objValue.Field(i), valmap, strings.ToLower(tag))
				}
				if len(key) == 0 {
					name = fieldName
				} else {
					name = key + "." + fieldName
				}
				// fmt.Printf("tag : %s\n", name)
				SetFields(objValue.Field(i), valmap, strings.ToLower(name))
				if err := FillStructRecursive(objValue.Field(i), valmap, name); err != nil {
					return err
				}
			}
		}
	case reflect.Ptr:
		if objValue.IsNil() {
			inst := reflect.New(objValue.Type().Elem())
			if err := FillStructRecursive(inst.Elem(), valmap, name); err != nil {
				return err
			}
			objValue.Set(inst)
		} else {
			if err := FillStructRecursive(objValue.Elem(), valmap, name); err != nil {
				return err
			}
		}
	default:
		return nil
	}
	// fmt.Printf("objValue end : %+v\n", objValue)
	return nil
}

//TO DO error
// SetFields sets value to fieldValue using tag as key in valmap
func SetFields(fieldValue reflect.Value, valmap map[string]interface{}, tag string) {
	if reflect.DeepEqual(fieldValue.Interface(), reflect.New(fieldValue.Type()).Elem().Interface()) {
		if fieldValue.CanSet() {
			if val, ok := valmap[tag]; ok {
				// fmt.Printf("tag %s : set %s in a %s\n", tag, val, fieldValue.Kind())
				fieldValue.Set(reflect.ValueOf(val).Elem().Convert(fieldValue.Type()))
			}
		} else {
			log.Fatalf("Error : type %s is not a settable ...\n", fieldValue.Type())
		}

	}

}
