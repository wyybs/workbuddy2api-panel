package upstream

import (
	"encoding/json"
	"strings"
)

// prepareKimiTools normalizes tool schemas for the CN Kimi outbound path.
// Other models retain their existing body, including its original bytes here.
func prepareKimiTools(body []byte, realm string) []byte {
	if realmKey(realm) != "cn" {
		return body
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(body, &obj) != nil {
		return body
	}
	var model string
	_ = json.Unmarshal(obj["model"], &model)
	if !strings.HasPrefix(strings.TrimPrefix(model, "cn:"), "kimi") {
		return body
	}
	var tools []map[string]any
	if json.Unmarshal(obj["tools"], &tools) != nil || len(tools) == 0 {
		return body
	}
	// A forced function selection must still name a declared tool.
	var choice struct {
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	_ = json.Unmarshal(obj["tool_choice"], &choice)
	kept := make([]map[string]any, 0, len(tools))
	budget := 10000
	for _, tool := range tools {
		fn, ok := tool["function"].(map[string]any)
		if !ok {
			// Preserve non-function tools rather than converting their protocol.
			kept = append(kept, tool)
			continue
		}
		name, _ := fn["name"].(string)
		if name != choice.Function.Name && (strings.HasPrefix(name, "mcp__rea__") || strings.HasPrefix(name, "mcp__tencent_docs__")) {
			continue
		}
		if params, ok := fn["parameters"].(map[string]any); ok {
			defs, _ := params["$defs"].(map[string]any)
			fn["parameters"] = inlineKimiSchemaBudget(params, defs, map[string]bool{}, 0, &budget)
			if budget <= 0 {
				return body
			}
		}
		kept = append(kept, tool)
	}
	encoded, err := json.Marshal(kept)
	if err != nil {
		return body
	}
	obj["tools"] = encoded
	out, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return out
}

// Cyclic schemas cannot be inlined; retain the local proxy's empty-schema
// fallback. Depth is bounded so an adversarial schema cannot exhaust the stack.
func inlineKimiSchema(node any, defs map[string]any, seen map[string]bool, depth int) any {
	budget := 10000
	return inlineKimiSchemaBudget(node, defs, seen, depth, &budget)
}

func inlineKimiSchemaBudget(node any, defs map[string]any, seen map[string]bool, depth int, budget *int) any {
	if *budget <= 0 {
		return nil
	}
	*budget--
	if depth > 64 {
		return map[string]any{}
	}
	switch v := node.(type) {
	case []any:
		out := make([]any, len(v))
		for i, x := range v {
			if *budget <= 0 {
				break
			}
			out[i] = inlineKimiSchemaBudget(x, defs, seen, depth+1, budget)
		}
		return out
	case map[string]any:
		if ref, ok := v["$ref"].(string); ok && strings.HasPrefix(ref, "#/$defs/") {
			name := strings.TrimPrefix(ref, "#/$defs/")
			name = strings.ReplaceAll(strings.ReplaceAll(name, "~1", "/"), "~0", "~")
			if seen[name] {
				return map[string]any{}
			}
			target, ok := defs[name]
			if !ok {
				return map[string]any{"type": "object"}
			}
			next := make(map[string]bool, len(seen)+1)
			for k, x := range seen {
				next[k] = x
			}
			next[name] = true
			return inlineKimiSchemaBudget(target, defs, next, depth+1, budget)
		}
		out := make(map[string]any, len(v))
		for k, x := range v {
			if *budget <= 0 {
				break
			}
			if k != "$defs" && k != "$schema" {
				out[k] = inlineKimiSchemaBudget(x, defs, seen, depth+1, budget)
			}
		}
		return out
	default:
		return node
	}
}
