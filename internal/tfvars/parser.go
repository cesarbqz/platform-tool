package tfvars

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func LoadTfvarsFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error loading tfvars file: %w", err)
	}

	return string(content), nil
}

func LoadValueFromMap(filePath, variableName, mapKey string) (string, error) {
	parser := hclparse.NewParser()

	// Parseamos el archivo .tfvars
	file, diags := parser.ParseHCLFile(filePath)
	if diags.HasErrors() {
		return "", fmt.Errorf("error parsing HCL file: %s", diags.Error())
	}

	// Obtenemos solo los atributos sin procesar todo el contenido
	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return "", fmt.Errorf("error retrieving attributes: %s", diags.Error())
	}

	// Buscamos solo el atributo que coincide con variableName
	attr, ok := attrs[variableName]
	if !ok {
		return "", fmt.Errorf("variable %s not found in the .tfvars file", variableName)
	}

	// Creamos un contexto de evaluación
	ctx := &hcl.EvalContext{}

	// Evaluamos el valor de la variable en el contexto HCL
	val, diags := attr.Expr.Value(ctx)
	if diags.HasErrors() {
		return "", fmt.Errorf("error retrieving value for variable %s: %s", variableName, diags.Error())
	}

	// Verificamos si el valor es de tipo mapa
	if !val.Type().IsObjectType() && !val.Type().IsMapType() {
		return "", fmt.Errorf("variable %s is not a map", variableName)
	}

	// Convertimos el valor en un mapa y buscamos la clave especificada
	mapVal := val.AsValueMap()
	entryVal, exists := mapVal[mapKey]
	if !exists {
		return "", fmt.Errorf("key %s not found in map %s", mapKey, variableName)
	}

	return entryVal.AsString(), nil
}

func ParseTfvars(filePath string) (map[string]interface{}, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading tfvars file: %w", err)
	}

	hclFile, diag := hclsyntax.ParseConfig(fileContent, filePath, hcl.Pos{Line: 1, Column: 1})
	if diag.HasErrors() {
		return nil, fmt.Errorf("error parsing tfvars file: %s", diag.Error())
	}

	variables := make(map[string]interface{})
	for name, expr := range hclFile.Body.(*hclsyntax.Body).Attributes {
		value, diag := expr.Expr.Value(nil)
		if diag.HasErrors() {
			return nil, fmt.Errorf("error evaluating variable '%s': %s", name, diag.Error())
		}

		goValue, err := convertCtyToGo(value)
		if err != nil {
			return nil, fmt.Errorf("error converting variable '%s': %s", name, err)
		}

		variables[name] = goValue
	}

	return variables, nil
}

func convertCtyToGo(value cty.Value) (interface{}, error) {
	if value.IsNull() {
		return nil, nil
	}

	switch {
	case value.Type().IsTupleType():
		return convertTupleToSlice(value), nil
	case value.Type().IsObjectType():
		return convertObjectToMap(value), nil
	case value.Type() == cty.String:
		return value.AsString(), nil
	case value.Type() == cty.Number:
		floatVal, _ := value.AsBigFloat().Float64()
		return floatVal, nil
	case value.Type() == cty.Bool:
		return value.True(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", value.Type().FriendlyName())
	}
}

func convertTupleToSlice(value cty.Value) []interface{} {
	if !value.CanIterateElements() {
		return nil
	}

	var result []interface{}
	it := value.ElementIterator()
	for it.Next() {
		_, elem := it.Element()
		goElem, _ := convertCtyToGo(elem)
		result = append(result, goElem)
	}
	return result
}

func convertObjectToMap(value cty.Value) map[string]interface{} {
	if !value.CanIterateElements() {
		return nil
	}

	result := make(map[string]interface{})
	it := value.ElementIterator()
	for it.Next() {
		key, val := it.Element()
		goVal, _ := convertCtyToGo(val)
		result[key.AsString()] = goVal
	}
	return result
}
