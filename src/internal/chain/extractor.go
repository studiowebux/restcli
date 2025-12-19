package chain

import (
	"encoding/json"
	"fmt"

	"github.com/jmespath/go-jmespath"
	"github.com/studiowebux/restcli/internal/types"
)

// ExtractVariables extracts variables from a response body using JMESPath expressions
func ExtractVariables(req *types.HttpRequest, responseBody string) (map[string]string, error) {
	if !HasExtractions(req) {
		return nil, nil
	}

	extracted := make(map[string]string)

	// Try to parse response as JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(responseBody), &jsonData); err != nil {
		// If not JSON, return error for extraction
		return nil, fmt.Errorf("cannot extract variables: response is not valid JSON")
	}

	// Extract each variable using JMESPath
	for varName, jmesPath := range req.Extract {
		result, err := jmespath.Search(jmesPath, jsonData)
		if err != nil {
			return nil, fmt.Errorf("failed to extract variable %s using path %s: %w", varName, jmesPath, err)
		}

		// Convert result to string
		var strValue string
		switch v := result.(type) {
		case string:
			strValue = v
		case float64:
			strValue = fmt.Sprintf("%g", v)
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%t", v)
		case nil:
			return nil, fmt.Errorf("variable %s: JMESPath %s returned null", varName, jmesPath)
		default:
			// For complex types, marshal to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("variable %s: failed to convert extracted value to string: %w", varName, err)
			}
			strValue = string(jsonBytes)
		}

		extracted[varName] = strValue
	}

	return extracted, nil
}
