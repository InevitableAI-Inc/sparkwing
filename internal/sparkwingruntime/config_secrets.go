package sparkwingruntime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sparkwing-dev/sparkwing/pkg/pipelines"
	"github.com/sparkwing-dev/sparkwing/sparkwing"
)

// WithPipelineSecrets installs the resolved Secrets struct on ctx.
func WithPipelineSecrets(ctx context.Context, v any) context.Context {
	return context.WithValue(ctx, sparkwing.RuntimePlumbing.PipelineSecrets, v)
}

// DecodePipelineConfig rehydrates a previously-resolved Config
// struct from a JSON blob. The struct's typed shape comes from the
// pipeline's Config() factory; the blob carries only values. Used
// by the cluster pod path to restore the typed Config the
// orchestrator-side resolution produced without re-running the
// yaml-layering logic on the pod.
//
// Returns (nil, nil) when the pipeline does not implement
// ConfigProvider or the blob is empty.
func DecodePipelineConfig(reg *sparkwing.Registration, raw []byte) (any, error) {
	if reg == nil || len(raw) == 0 {
		return nil, nil
	}
	inst := reg.Instance()
	if inst == nil {
		return nil, nil
	}
	cp, ok := inst.(sparkwing.ConfigProvider)
	if !ok {
		return nil, nil
	}
	cfg := cp.Config()
	if cfg == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(cfg)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("pipeline %q config: Config() must return a non-nil pointer to a struct, got %T", reg.Name, cfg)
	}
	if err := json.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("pipeline %q config: decode persisted blob: %w", reg.Name, err)
	}
	return cfg, nil
}

// ResolvePipelineSecrets resolves every required secret declared by
// the pipeline -- both via the SecretsProvider's struct fields and
// via the SecretsField list on the yaml entry -- against the
// SecretResolver installed on ctx, before the pipeline's Plan runs.
//
// Required fail-fast: a missing required secret produces a clear
// error naming the pipeline and the secret. Optional entries whose
// resolver error matches ErrSecretMissing leave the struct field
// empty without failing the run; other errors propagate.
//
// Resolved values populate the Secrets struct fields by sw-tag name,
// so step bodies can read sec.DeployToken directly via
// PipelineSecrets[T](ctx) without re-fetching.
//
// Returns nil, nil when the pipeline value does not implement
// SecretsProvider and the yaml entry declares no required secrets --
// nothing to install. Returns the populated struct (or a synthesized
// zero struct when only the yaml side declares secrets) otherwise.
func ResolvePipelineSecrets(ctx context.Context, reg *sparkwing.Registration, yamlEntry *pipelines.Pipeline) (any, error) {
	if reg == nil {
		return nil, nil
	}
	p := reg.Instance()
	if p == nil {
		return nil, nil
	}
	resolver, _ := ctx.Value(sparkwing.RuntimePlumbing.SecretResolver).(sparkwing.SecretResolver)

	// SecretsField from yaml: union with the struct-declared required set.
	var yamlRequired []string
	var yamlOptional []string
	if yamlEntry != nil {
		for _, e := range yamlEntry.Secrets {
			if e.IsRequired() {
				yamlRequired = append(yamlRequired, e.Name)
			} else {
				yamlOptional = append(yamlOptional, e.Name)
			}
		}
	}

	sp, hasProvider := p.(sparkwing.SecretsProvider)
	if !hasProvider && len(yamlRequired) == 0 && len(yamlOptional) == 0 {
		return nil, nil
	}

	var sec any
	var specs []swFieldSpec
	var elem reflect.Value
	if hasProvider {
		sec = sp.Secrets()
		if sec != nil {
			rv := reflect.ValueOf(sec)
			if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
				return nil, fmt.Errorf("pipeline %q secrets: Secrets() must return a non-nil pointer to a struct, got %T", reg.Name, sec)
			}
			ss, err := parseSWTags(rv.Type())
			if err != nil {
				return nil, fmt.Errorf("pipeline %q secrets: %w", reg.Name, err)
			}
			specs = ss
			elem = rv.Elem()
			for i := range specs {
				// Secrets default to required when neither flag is set,
				// matching the bare-string SecretsField rule.
				if !specs[i].required && !specs[i].optional {
					specs[i].required = true
				}
			}
		}
	}

	// Build the union of names to resolve, tracking required-ness per
	// name. Struct entries win on conflict because they carry the
	// destination field.
	requiredNames := map[string]struct{}{}
	optionalNames := map[string]struct{}{}
	for _, n := range yamlRequired {
		requiredNames[n] = struct{}{}
	}
	for _, n := range yamlOptional {
		if _, alreadyReq := requiredNames[n]; alreadyReq {
			continue
		}
		optionalNames[n] = struct{}{}
	}
	for _, s := range specs {
		if s.required {
			requiredNames[s.name] = struct{}{}
			delete(optionalNames, s.name)
		} else if s.optional {
			if _, alreadyReq := requiredNames[s.name]; !alreadyReq {
				optionalNames[s.name] = struct{}{}
			}
		}
	}

	if (len(requiredNames) > 0 || len(optionalNames) > 0) && resolver == nil {
		return nil, fmt.Errorf("pipeline %q secrets: declared but no SecretResolver installed on ctx", reg.Name)
	}

	// Resolve every name, populate the struct field when one exists.
	specByName := map[string]swFieldSpec{}
	for _, s := range specs {
		specByName[s.name] = s
	}

	// Required first so the run fails before any optional resolution.
	for name := range requiredNames {
		v, _, err := resolver.Resolve(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("pipeline %q secrets: %q: %w", reg.Name, name, err)
		}
		if s, ok := specByName[name]; ok {
			if err := coerceAssign(elem.FieldByIndex(s.field.Index), v, s.field.Name); err != nil {
				return nil, fmt.Errorf("pipeline %q secrets: %w", reg.Name, err)
			}
		}
	}
	for name := range optionalNames {
		v, _, err := resolver.Resolve(ctx, name)
		if err != nil {
			if errors.Is(err, sparkwing.ErrSecretMissing) {
				continue
			}
			return nil, fmt.Errorf("pipeline %q secrets: %q: %w", reg.Name, name, err)
		}
		if s, ok := specByName[name]; ok {
			if err := coerceAssign(elem.FieldByIndex(s.field.Index), v, s.field.Name); err != nil {
				return nil, fmt.Errorf("pipeline %q secrets: %w", reg.Name, err)
			}
		}
	}

	return sec, nil
}

// swFieldSpec captures the parsed sw + default tags on one struct
// field. Mirrors sparkwing.swFieldSpec; duplicated here so this
// package can move without exporting sparkwing-internal types.
type swFieldSpec struct {
	field      reflect.StructField
	name       string
	required   bool
	optional   bool
	defaultRaw string
	hasDefault bool
}

func parseSWTags(t reflect.Type) ([]swFieldSpec, error) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}
	out := make([]swFieldSpec, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		raw, ok := f.Tag.Lookup("sw")
		if !ok {
			continue
		}
		parts := strings.Split(raw, ",")
		name := strings.TrimSpace(parts[0])
		if name == "" {
			return nil, fmt.Errorf("field %s: sw tag has empty name", f.Name)
		}
		spec := swFieldSpec{field: f, name: name}
		for _, mod := range parts[1:] {
			switch strings.TrimSpace(mod) {
			case "required":
				spec.required = true
			case "optional":
				spec.optional = true
			case "":
				// trailing comma, ignore
			default:
				return nil, fmt.Errorf("field %s: unknown sw modifier %q", f.Name, mod)
			}
		}
		if spec.required && spec.optional {
			return nil, fmt.Errorf("field %s: sw tag cannot set both required and optional", f.Name)
		}
		if d, has := f.Tag.Lookup("default"); has {
			spec.defaultRaw = d
			spec.hasDefault = true
			if spec.required {
				return nil, fmt.Errorf("field %s: required and default cannot both be set", f.Name)
			}
		}
		out = append(out, spec)
	}
	return out, nil
}

func coerceAssign(fv reflect.Value, raw any, fieldName string) error {
	if !fv.CanSet() {
		return fmt.Errorf("field %s: cannot set", fieldName)
	}
	if raw == nil {
		return nil
	}
	switch fv.Kind() {
	case reflect.String:
		s, err := toString(raw)
		if err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
		fv.SetString(s)
	case reflect.Bool:
		b, err := toBool(raw)
		if err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(raw)
		if err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
		if fv.OverflowInt(n) {
			return fmt.Errorf("field %s: value %d overflows %s", fieldName, n, fv.Type())
		}
		fv.SetInt(n)
	case reflect.Float32, reflect.Float64:
		x, err := toFloat64(raw)
		if err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
		fv.SetFloat(x)
	default:
		rv := reflect.ValueOf(raw)
		if rv.Type().AssignableTo(fv.Type()) {
			fv.Set(rv)
			return nil
		}
		return fmt.Errorf("field %s: cannot assign %T to %s", fieldName, raw, fv.Type())
	}
	return nil
}

func toString(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case int:
		return strconv.Itoa(t), nil
	case int64:
		return strconv.FormatInt(t, 10), nil
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64), nil
	case bool:
		return strconv.FormatBool(t), nil
	}
	return "", fmt.Errorf("expected string-compatible value, got %T", v)
}

func toBool(v any) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		return strconv.ParseBool(t)
	}
	return false, fmt.Errorf("expected bool, got %T", v)
}

func toInt64(v any) (int64, error) {
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint, uint8, uint16, uint32, uint64:
		rv := reflect.ValueOf(t)
		return int64(rv.Uint()), nil
	case float64:
		if t != float64(int64(t)) {
			return 0, fmt.Errorf("expected integer, got fractional %v", t)
		}
		return int64(t), nil
	case string:
		return strconv.ParseInt(t, 10, 64)
	}
	return 0, fmt.Errorf("expected integer, got %T", v)
}

func toFloat64(v any) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	}
	return 0, fmt.Errorf("expected float, got %T", v)
}
