// Package posflag implements a koanf.Provider that reads commandline
// parameters as conf maps using spf13/pflag, a POSIX compliant
// alternative to Go's stdlib flag package.
package posflag

import (
	"errors"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/maps"
	"github.com/spf13/pflag"
)

// Posflag implements a pflag command line provider.
type Posflag struct {
	delim           string
	flagset         *pflag.FlagSet
	ko              *koanf.Koanf
	valueCallback   func(key string, value string) (string, interface{})
	keyNameCallback func(flag *pflag.Flag) string
}

// Option configures some aspect of Posflag provider.
type Option func(*Posflag)

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
func Provider(f *pflag.FlagSet, delim string, ko *koanf.Koanf) *Posflag {
	return &Posflag{
		flagset: f,
		delim:   delim,
		ko:      ko,
	}
}

// ProviderWithValue works exactly the same as Provider except the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
// Deprecated: this function is deprecated, use WithValue option to achieve the same result
func ProviderWithValue(f *pflag.FlagSet, delim string, ko *koanf.Koanf, cb func(key string, value string) (string, interface{})) *Posflag {
	return &Posflag{
		flagset:       f,
		delim:         delim,
		ko:            ko,
		valueCallback: cb,
	}
}

// ProviderWithOptions works exactly the same as Provider except for the optional configuration callbacks that provide
// extendable way to configure Posflag without changing of method signature
func ProviderWithOptions(f *pflag.FlagSet, delim string, options ...Option) *Posflag {
	p := &Posflag{
		flagset: f,
		delim:   delim,
	}
	if options != nil {
		for _, opt := range options {
			opt(p)
		}
	}
	return p
}

// Read reads the flag variables and returns a nested conf map.
func (p *Posflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	p.flagset.VisitAll(func(f *pflag.Flag) {
		keyName := f.Name
		if p.keyNameCallback != nil {
			keyName = p.keyNameCallback(f)
		}
		// If no value was explicitly set in the command line,
		// check if the default value should be used.
		if !f.Changed {
			if p.ko != nil {
				if p.ko.Exists(keyName) {
					return
				}
			} else {
				return
			}
		}

		var v interface{}
		switch f.Value.Type() {
		case "int":
			i, _ := p.flagset.GetInt(f.Name)
			v = int64(i)
		case "int8":
			i, _ := p.flagset.GetInt8(f.Name)
			v = int64(i)
		case "int16":
			i, _ := p.flagset.GetInt16(f.Name)
			v = int64(i)
		case "int32":
			i, _ := p.flagset.GetInt32(f.Name)
			v = int64(i)
		case "int64":
			i, _ := p.flagset.GetInt64(f.Name)
			v = int64(i)
		case "float32":
			v, _ = p.flagset.GetFloat32(f.Name)
		case "float":
			v, _ = p.flagset.GetFloat64(f.Name)
		case "bool":
			v, _ = p.flagset.GetBool(f.Name)
		case "stringSlice":
			v, _ = p.flagset.GetStringSlice(f.Name)
		case "intSlice":
			v, _ = p.flagset.GetIntSlice(f.Name)
		default:
			if p.valueCallback != nil {
				key, value := p.valueCallback(keyName, f.Value.String())
				if key == "" {
					return
				}
				mp[key] = value
				return
			} else {
				v = f.Value.String()
			}
		}

		mp[keyName] = v
	})
	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the env koanf.
func (p *Posflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("pflag provider does not support this method")
}

// Watch is not supported.
func (p *Posflag) Watch(cb func(event interface{}, err error)) error {
	return errors.New("posflag provider does not support this method")
}

// ParentKoanf option adds the Koanf instance to see if the
// the flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func ParentKoanf(ko *koanf.Koanf) Option {
	return func(p *Posflag) {
		p.ko = ko
	}
}

// ValueCallback options adds the callback
// takes a (key, value) with the variable name and value and allows you
// to modify both. This is useful for cases where you may want to return
// other types like a string slice instead of just a string.
func ValueCallback(cb func(key string, value string) (string, interface{})) Option {
	return func(p *Posflag) {
		p.valueCallback = cb
	}
}

// RenameCallback options adds the possibility to map flags in case when flag name
// differs from setting name.
func RenameCallback(cb func(flag *pflag.Flag) string) Option {
	return func(p *Posflag) {
		p.keyNameCallback = cb
	}
}
