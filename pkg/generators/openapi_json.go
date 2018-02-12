package generators

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	"k8s.io/gengo/parser"
	"k8s.io/gengo/types"
	openapi "k8s.io/kube-openapi/pkg/common"
)

func NewOpenAPIGenJSON(output io.Writer, targetPkg, targetType string) *openAPIGeneratorJSON {
	return &openAPIGeneratorJSON{output, targetPkg, targetType}
}

// openAPIGeneratorJSON is a custom generator implementation that builds an
// openapi spec structure and writes it out as JSON or YAML. It reuses the
// gengo parser package for parsing types, but otherwise is a custom
// implementation due to gengo being limited to generating Go code.
type openAPIGeneratorJSON struct {
	output                io.Writer
	targetPkg, targetType string
}

func (g openAPIGeneratorJSON) Execute() error {
	b := parser.New()
	if err := b.AddDir(g.targetPkg); err != nil {
		return err
	}
	u, err := b.FindTypes()
	if err != nil {
		return err
	}
	pkg := u.Package(g.targetPkg)
	t := pkg.Type(g.targetType)
	return g.generate(t)
}

func (g openAPIGeneratorJSON) generate(t *types.Type) error {
	schema := &spec.Schema{}
	// Only generate for struct type and ignore the rest
	switch t.Kind {
	case types.Struct:
		// TODO: support manually defined openapi definitions.
		// This really would involve compiling the code in question and calling
		// the function.
		if hasOpenAPIDefinitionMethod(t) {
			return nil
		}
		props := spec.SchemaProps{
			Description: strings.TrimSpace(parseDescription(t.CommentLines)),
		}
		properties, req, err := g.generateMembers(t, true)
		if err != nil {
			return err
		}
		props.Properties = properties
		props.Required = req
		schema.SchemaProps = props

		extensions, err := g.generateExtensions(t.CommentLines)
		if err != nil {
			return err
		}
		schema.VendorExtensible = *extensions
	}
	return json.NewEncoder(g.output).Encode(schema)
}

func (g openAPIGeneratorJSON) generateMembers(t *types.Type, root bool) (map[string]spec.Schema, []string, error) {
	props := map[string]spec.Schema{}
	var required []string
	for _, m := range t.Members {
		name := getReferableName(&m)
		if root && name != "spec" && name != "status" {
			continue
		}
		if hasOpenAPITagValue(m.CommentLines, tagValueFalse) {
			continue
		}
		if shouldInlineMembers(&m) {
			memberSchema, memberRequired, err := g.generateMembers(m.Type, false)
			if err != nil {
				return nil, nil, err
			}
			for k, v := range memberSchema {
				props[k] = v
			}
			required = append(required, memberRequired...)
			continue
		}
		if name == "" {
			continue
		}
		schema, err := g.generateProperty(&m, t)
		if err != nil {
			return nil, nil, err
		}
		glog.V(2).Infof("Member %q has %d properties", name, len((*schema).Properties))
		props[name] = *schema
	}
	glog.V(2).Infof("Type being generated has %d properties", len(props))
	return props, required, nil
}

func (g openAPIGeneratorJSON) generateProperty(m *types.Member, parent *types.Type) (*spec.Schema, error) {
	var err error
	if err := validatePatchTags(m, parent); err != nil {
		return nil, err
	}
	extensions, err := g.generateExtensions(m.CommentLines)
	if err != nil {
		return nil, err
	}
	t := resolveAliasAndPtrType(m.Type)
	jsonTags := getJsonTags(m)
	if len(jsonTags) > 1 && jsonTags[1] == "string" {
		schema := &spec.Schema{}
		schema.VendorExtensible = *extensions
		schema.SchemaProps.Type = []string{"string"}
		return schema, nil
	}
	// If we can get a openAPI type and format for this type, we consider it to be simple property
	typeString, format := openapi.GetOpenAPITypeFormat(t.String())
	if typeString != "" {
		// we don't call getSimpleProperty here else we'd overwrite the
		// extensions meta above
		schema := g.generateSimpleProperty(typeString, format)
		schema.VendorExtensible = *extensions
		return schema, nil
	}
	schema := &spec.Schema{}
	switch t.Kind {
	case types.Builtin:
		return nil, fmt.Errorf("please add type %v to getOpenAPITypeFormat function", t)
	case types.Map:
		schema, err = g.generateMapProperty(t)
		if err != nil {
			return nil, err
		}
	case types.Slice, types.Array:
		schema, err = g.generateSliceProperty(t)
		if err != nil {
			return nil, err
		}
	case types.Struct, types.Interface:
		props, req, err := g.generateMembers(t, false)
		if err != nil {
			return nil, err
		}
		schema.Properties = props
		schema.SchemaProps.Required = req
	default:
		return nil, fmt.Errorf("cannot generate spec for type %v", t)
	}
	schema.SchemaProps.Description = strings.TrimSpace(parseDescription(m.CommentLines))
	schema.VendorExtensible = *extensions
	return schema, nil
}

func (g openAPIGeneratorJSON) generateMapProperty(t *types.Type) (*spec.Schema, error) {
	var err error
	keyType := resolveAliasAndPtrType(t.Key)
	elemType := resolveAliasAndPtrType(t.Elem)

	// According to OpenAPI examples, only map from string is supported
	if keyType.Name.Name != "string" {
		return nil, fmt.Errorf("map with non-string keys are not supported by OpenAPI in %v", t)
	}
	typeString, format := openapi.GetOpenAPITypeFormat(elemType.String())
	if typeString != "" {
		schema := &spec.Schema{}
		schema.Type = []string{"object"}
		schema.AdditionalProperties = &spec.SchemaOrBool{
			Schema: g.generateSimpleProperty(typeString, format),
		}
		return schema, nil
	}

	// we don't set any fields on this variable until after the switch so that
	// we can easily call `schema, err = g.generateXYZ()` and set the fields we
	// need at the end
	schema := &spec.Schema{}
	switch elemType.Kind {
	case types.Builtin:
		return nil, fmt.Errorf("please add type %v to getOpenAPITypeFormat function", elemType)
	case types.Struct:
		props, req, err := g.generateMembers(elemType, false)
		if err != nil {
			return nil, err
		}
		schema.AdditionalProperties = &spec.SchemaOrBool{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: props,
					Required:   req,
				},
			},
		}
	case types.Slice, types.Array:
		schema, err = g.generateSliceProperty(elemType)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("map Element kind %v is not supported in %v", elemType.Kind, t.Name)
	}
	schema.Type = []string{"object"}
	return schema, nil
}

func (g openAPIGeneratorJSON) generateSliceProperty(t *types.Type) (*spec.Schema, error) {
	elemType := resolveAliasAndPtrType(t.Elem)
	typeString, format := openapi.GetOpenAPITypeFormat(elemType.String())
	if typeString != "" {
		schema := &spec.Schema{}
		schema.Type = []string{"array"}
		schema.AdditionalProperties = &spec.SchemaOrBool{
			Schema: g.generateSimpleProperty(typeString, format),
		}
		return schema, nil
	}
	schema := &spec.Schema{}
	switch elemType.Kind {
	case types.Builtin:
		return nil, fmt.Errorf("please add type %v to getOpenAPITypeFormat function", elemType)
	case types.Struct:
		props, req, err := g.generateMembers(elemType, false)
		if err != nil {
			return nil, err
		}
		glog.V(1).Infof("Got member props: %+v", props)
		schema.AdditionalProperties = &spec.SchemaOrBool{Schema: &spec.Schema{}}
		schema.AdditionalProperties.Schema.Properties = props
		schema.AdditionalProperties.Schema.Required = req
	default:
		return nil, fmt.Errorf("slice Element kind %v is not supported in %v", elemType.Kind, t)
	}
	schema.Type = []string{"array"}
	return schema, nil
}

func (g openAPIGeneratorJSON) generateSimpleProperty(typeString, format string) *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:   []string{typeString},
			Format: format,
		},
	}
}

func (g openAPIGeneratorJSON) generateExtensions(CommentLines []string) (*spec.VendorExtensible, error) {
	tagValues := getOpenAPITagValue(CommentLines)
	type NameValue struct {
		Name, Value string
	}
	extensions := []NameValue{}
	for _, val := range tagValues {
		if strings.HasPrefix(val, tagExtensionPrefix) {
			parts := strings.SplitN(val, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid extension value: %v", val)
			}
			extensions = append(extensions, NameValue{parts[0], parts[1]})
		}
	}
	patchMergeKeyTag, err := getSingleTagsValue(CommentLines, tagPatchMergeKey)
	if err != nil {
		return nil, err
	}
	if len(patchMergeKeyTag) > 0 {
		extensions = append(extensions, NameValue{tagExtensionPrefix + patchMergeKeyExtensionName, patchMergeKeyTag})
	}
	patchStrategyTag, err := getSingleTagsValue(CommentLines, tagPatchStrategy)
	if err != nil {
		return nil, err
	}
	if len(patchStrategyTag) > 0 {
		extensions = append(extensions, NameValue{tagExtensionPrefix + patchStrategyExtensionName, patchStrategyTag})
	}
	vE := &spec.VendorExtensible{
		Extensions: map[string]interface{}{},
	}
	for _, extension := range extensions {
		vE.Extensions[extension.Name] = extension.Value
	}
	return vE, nil
}
