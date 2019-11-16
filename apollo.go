package apollo

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

func WatchConfig(config interface{}) error {
	updater, err := newConfigUpdater(config)
	if err != nil {
		return err
	}

	keys := GetAllKeys(defaultNamespace)
	for _, key := range keys {
		fieldMeta, ok := updater.fieldsMeta[key]
		if !ok {
			continue
		}

		val := GetStringValue(key, fieldMeta.apolloDefault)
		if err := updater.setValue(fieldMeta.fieldName, val); err != nil {
			return err
		}
	}

	events := WatchUpdate()
	go func() {
		for {
			select {
			case event := <-events:
				allChanges := make(map[string]string)
				for key, change := range event.Changes {
					allChanges[key] = change.NewValue
				}

				go func() {
					if err := updater.Update(allChanges); err != nil {
						log.Printf("error update changes: %v", allChanges)
					}
				}()
			}
		}
	}()

	return nil
}

type fieldMeta struct {
	fieldName      string
	apolloKey      string
	apolloCallback string
	apolloDefault  string
}

type configUpdater struct {
	config         interface{}
	configType     reflect.Type
	configElemType reflect.Type
	configVal      reflect.Value
	configElemVal  reflect.Value
	fieldsMeta     map[string]fieldMeta
}

func newConfigUpdater(v interface{}) (*configUpdater, error) {
	configType := reflect.TypeOf(v)
	configElemType := configType
	configVal := reflect.ValueOf(v)
	configElemVal := configVal
	if configType.Kind() == reflect.Ptr {
		configElemType = configType.Elem()
		configElemVal = configVal.Elem()
	}
	if configElemType.Kind() != reflect.Struct {
		return nil, errors.New("invalid config")
	}

	config := &configUpdater{
		v,
		configType,
		configElemType,
		configVal,
		configElemVal,
		make(map[string]fieldMeta),
	}
	if err := config.parserConfig(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *configUpdater) Update(kv map[string]string) error {
	oldConfig := elem(copyStruct(c.config))
	methods := map[string]struct{}{}
	for k, v := range kv {
		configField, ok := c.fieldsMeta[k]
		if !ok {
			continue
		}

		if err := c.setValue(configField.fieldName, v); err != nil {
			return err
		}

		methods[configField.apolloCallback] = struct{}{}
	}

	for method := range methods {
		if err := c.callMethod(method, oldConfig); err != nil {
			return err
		}
	}

	return nil
}

func (c *configUpdater) parserConfig() error {
	for i := 0; i < c.configElemType.NumField(); i++ {
		fieldElemType := c.configElemType.Field(i)
		if isLower(fieldElemType.Name) {
			continue
		}

		apolloKey := fieldElemType.Tag.Get("apollo_key")
		if apolloKey == "" {
			apolloKey = fieldElemType.Name
		}
		apolloCallback := fieldElemType.Tag.Get("apollo_callback")
		apolloDefault := fieldElemType.Tag.Get("apollo_default")

		if apolloCallback != "" {
			if _, ok := c.configType.MethodByName(apolloCallback); !ok {
				return fmt.Errorf("invalid callback: %s", apolloCallback)
			}
		}

		configField := fieldMeta{
			fieldName:      fieldElemType.Name,
			apolloKey:      apolloKey,
			apolloCallback: apolloCallback,
			apolloDefault:  apolloDefault,
		}
		c.fieldsMeta[apolloKey] = configField
	}

	return nil
}

func (c *configUpdater) callMethod(methodName string, methodArgs ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	method := c.configVal.MethodByName(methodName)
	if !method.IsValid() {
		return errors.New("method not exist")
	}

	var args []reflect.Value
	for _, arg := range methodArgs {
		args = append(args, reflect.ValueOf(arg))
	}
	method.Call(args)

	return nil
}

func (c *configUpdater) setValue(fieldName, newValue string) error {
	field := c.configElemVal.FieldByName(fieldName)
	if !field.CanSet() {
		return errors.New("field cannot be set")
	}
	isEmpty := strings.TrimSpace(newValue) == ""

	switch field.Kind() {
	case reflect.String:
		field.Set(reflect.ValueOf(newValue))
	case reflect.Int:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.Atoi(newValue)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(i))
	case reflect.Int8:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseInt(newValue, 10, 8)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(int8(i)))
	case reflect.Int16:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseInt(newValue, 10, 16)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(int16(i)))
	case reflect.Int32:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseInt(newValue, 10, 32)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(int32(i)))
	case reflect.Int64:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseInt(newValue, 10, 64)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(i))
	case reflect.Uint:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseUint(newValue, 10, 0)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(uint(i)))
	case reflect.Uint8:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseUint(newValue, 10, 8)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(uint8(i)))
	case reflect.Uint16:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseUint(newValue, 10, 16)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(uint16(i)))
	case reflect.Uint32:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseUint(newValue, 10, 32)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(uint32(i)))
	case reflect.Uint64:
		if isEmpty {
			newValue = "0"
		}
		i, err := strconv.ParseUint(newValue, 10, 64)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(i))
	case reflect.Bool:
		if isEmpty {
			newValue = "false"
		}
		b, err := strconv.ParseBool(newValue)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(b))
	case reflect.Float32:
		if isEmpty {
			newValue = "0"
		}
		f, err := strconv.ParseFloat(newValue, 32)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(float32(f)))
	case reflect.Float64:
		if isEmpty {
			newValue = "0"
		}
		f, err := strconv.ParseFloat(newValue, 64)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(f))
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}
