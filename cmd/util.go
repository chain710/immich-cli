package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/chain710/immich-cli/client"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func newClient() client.ClientWithResponsesInterface {
	api := viper.GetString(ViperKey_API)
	key := viper.GetString(ViperKey_APIKey)

	options := []client.ClientOption{
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-Api-Key", key)
			return nil
		}),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			log.DebugFn(func() []interface{} {
				return []any{
					fmt.Sprintf("%s: `%s`", req.Method, req.URL.String()),
					fmt.Sprintf("Header: `%v`", req.Header),
				}
			})
			return nil
		}),
	}

	cli, err := client.NewClientWithResponses(api, options...)
	if err != nil {
		log.Fatalf("create immich client error: %v", err)
	}

	return cli
}

func parseOptions(s string) []string {
	var ss []string
	segments := strings.Split(s, ",")
	for _, segment := range segments {
		ss = append(ss, strings.TrimSpace(segment))
	}
	return ss
}

func addFlagSetByFormFields(s any, set *pflag.FlagSet) {
	typeInfo := resolveElem(reflect.ValueOf(s)).Type()
	for i := 0; i < typeInfo.NumField(); i++ {
		fieldType := typeInfo.Field(i)
		optionsRaw, ok := fieldType.Tag.Lookup("form")
		if !ok {
			continue
		}
		options := parseOptions(optionsRaw)
		if len(options) == 0 {
			continue
		}

		name := options[0]
		set.Var(&genericVar{t: fieldType.Type.String()}, name, "")
	}
}

func setFormFields(s any, set *pflag.FlagSet) error {
	tagFields, err := validateAndGetFieldMap(s)
	if err != nil {
		return err
	}

	var errs []error
	set.VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}

		fieldValue, ok := tagFields[flag.Name]
		if !ok {
			log.Debugf("no flag `%s` in params field", flag.Name)
			return
		}

		log.Debugf("ready to set value by flag: %s, value: %s", flag.Name, flag.Value.String())
		switch fieldValue.Kind() {
		case reflect.Pointer:
			if err := setPointerField(fieldValue, flag.Value.String()); err != nil {
				errs = append(errs, err)
			}
		default:
			errs = append(errs, fmt.Errorf("unsupport field type %s", fieldValue.Type().String()))
		}
	})

	return errors.Join(errs...)
}

func resolveElem(value reflect.Value) reflect.Value {
	if value.Type().Kind() == reflect.Pointer || value.Type().Kind() == reflect.Interface {
		return resolveElem(value.Elem())
	}

	return value
}

func setPointerField(ptrField reflect.Value, value string) error {
	elemType := ptrField.Type().Elem()
	newValue := reflect.New(elemType) // NOTE: newValue is a pointer type
	ptrField.Set(newValue)
	switch elemType.Kind() {
	case reflect.Bool:
		if val, err := strconv.ParseBool(value); err != nil {
			return err
		} else {
			newValue.Elem().SetBool(val)
		}
	case reflect.Float32:
		if val, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		} else {
			newValue.Elem().SetFloat(val)
		}
	case reflect.String:
		newValue.Elem().SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(value, 10, 64); err != nil {
			return err
		} else {
			newValue.Elem().SetInt(val)
		}
	default:
		switch newValue.Interface().(type) {
		case *uuid.UUID:
			if id, err := uuid.Parse(value); err != nil {
				return err
			} else {
				newValue.Elem().Set(reflect.ValueOf(id))
			}
		default:
			return fmt.Errorf("unsupported type: %s", elemType.String())
		}
	}

	return nil
}

// validateAndGetFieldMap make sure every field of s is ptr to a valid primitive type or struct
// return a map, which key is tag form name, value is field's reflect.Value
func validateAndGetFieldMap(s any) (map[string]reflect.Value, error) {
	valueInfo := resolveElem(reflect.ValueOf(s))
	typeInfo := valueInfo.Type()
	ret := make(map[string]reflect.Value)
	for i := 0; i < valueInfo.NumField(); i++ {
		fieldValue := valueInfo.Field(i)
		fieldType := typeInfo.Field(i)
		optionsRaw, ok := fieldType.Tag.Lookup("form")
		if !ok {
			continue
		}
		options := parseOptions(optionsRaw)
		if len(options) == 0 {
			return nil, fmt.Errorf("malform field `%s`, tag options is empty", fieldType.Name)
		}

		name := options[0]
		switch fieldType.Type.Kind() {
		case reflect.Pointer, reflect.Array, reflect.Slice:
		default:
			return nil, fmt.Errorf("malform field `%s`, type: %s", fieldType.Name, fieldType.Type.String())
		}

		switch fieldType.Type.Elem().Kind() {
		case reflect.Interface,
			reflect.UnsafePointer,
			reflect.Pointer:
			return nil, fmt.Errorf("malform field `%s`, elem type[%d]: %s",
				fieldType.Name, fieldType.Type.Elem().Kind(), fieldType.Type.Elem().String())
		}

		ret[name] = fieldValue
	}

	return ret, nil
}
