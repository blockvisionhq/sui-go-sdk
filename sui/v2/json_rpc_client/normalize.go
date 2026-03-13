// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json_rpc_client

import (
	"regexp"
	"strings"
)

var (
	coinTypeRe = regexp.MustCompile(`0x([0-9a-fA-F]{1,63})(::|,|<|>|$)`)
	// Parse JSON RPC VecMap debug format: key: X \n value: Y
	vecMapKeyRe   = regexp.MustCompile(`key:\s*([^\n,]+)`)
	vecMapValueRe = regexp.MustCompile(`value:\s*([^\n]+)`)
)

// normalizeObjectJSON normalizes object json (Content.Fields primary), recursively flattens Move structure to match gRPC
func normalizeObjectJSON(jsonMap map[string]interface{}, objectId string) map[string]interface{} {
	if jsonMap == nil {
		return nil
	}
	flattened := flattenMoveValue(jsonMap)
	out, _ := flattened.(map[string]interface{})
	if out == nil {
		out = make(map[string]interface{})
		for k, v := range jsonMap {
			out[k] = v
		}
	}
	// id: UID object {"id":"0x..."} to string, or use objectId when missing
	if idVal, ok := out["id"]; ok && idVal != nil {
		if m, ok := idVal.(map[string]interface{}); ok {
			if sid, ok := m["id"].(string); ok && sid != "" {
				out["id"] = sid
			}
		}
	}
	if _, has := out["id"]; !has && objectId != "" {
		out["id"] = objectId
	}
	// Recursively normalize attributes that are string (Display VecMap debug format)
	normalizeAttributesRecursive(out)

	// Recursively flatten nested UID objects {"id":"0x..."} -> "0x..."
	flattenUIDRecursive(out)

	// Fix duplicated prefix in link
	if link, ok := out["link"].(string); ok && link != "" {
		if fixed := fixDuplicatedLinkPrefix(link); fixed != "" && fixed != link {
			out["link"] = fixed
		}
	}
	return out
}

// flattenMoveValue recursively flattens Move Content structure: {fields:X, type:T} -> flatten(X)
// VecMap: {fields:{contents:[{fields:{key,value},type},...]},type} -> {contents:[{key,value},...]}
// TypeName: {fields:{name:"..."},type} -> {name:"..."}
func flattenMoveValue(v interface{}) interface{} {
	switch val := v.(type) {
	case nil:
		return nil
	case string, float64, bool:
		return val
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, e := range val {
			out[i] = flattenMoveValue(e)
		}
		return out
	case map[string]interface{}:
		// Move struct: {fields: X, type: T} -> replace with flatten(X)
		if fields, ok := val["fields"]; ok {
			ftype, _ := val["type"].(string)
			// VecMap: fields.contents is Entry array
			if fmap, ok := fields.(map[string]interface{}); ok {
				if arr, ok := fmap["contents"].([]interface{}); ok {
					if strings.Contains(ftype, "VecMap") {
						contents := make([]map[string]string, 0, len(arr))
						for _, e := range arr {
							entry, ok := e.(map[string]interface{})
							if !ok {
								continue
							}
							ef, ok := entry["fields"].(map[string]interface{})
							if !ok {
								continue
							}
							k, _ := ef["key"].(string)
							v, _ := ef["value"].(string)
							contents = append(contents, map[string]string{"key": k, "value": v})
						}
						if len(contents) > 0 {
							return map[string]interface{}{"contents": contents}
						}
					}
				}
				// TypeName: fields.name
				if strings.Contains(ftype, "TypeName") {
					if nm, ok := fmap["name"].(string); ok {
						return map[string]interface{}{"name": nm}
					}
				}
				// Plain struct: recursively flatten its fields
				flat := make(map[string]interface{})
				for k, fv := range fmap {
					flat[k] = flattenMoveValue(fv)
				}
				return flat
			}
		}
		// Already flatten format (no fields/type wrapper) or gRPC format, recursively process children
		out := make(map[string]interface{})
		for k, fv := range val {
			out[k] = flattenMoveValue(fv)
		}
		return out
	default:
		return val
	}
}

// flattenUIDRecursive recursively flattens UID objects: {"id":"0x..."} -> "0x..." to match gRPC
func flattenUIDRecursive(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, fv := range val {
			if uid, ok := fv.(map[string]interface{}); ok && len(uid) == 1 {
				if sid, ok := uid["id"].(string); ok && len(sid) > 10 && strings.HasPrefix(sid, "0x") {
					val[k] = sid
					continue
				}
			}
			flattenUIDRecursive(fv)
		}
	case []interface{}:
		for _, e := range val {
			flattenUIDRecursive(e)
		}
	}
}

// normalizeAttributesRecursive recursively converts attributes strings to gRPC format
func normalizeAttributesRecursive(m map[string]interface{}) {
	for k, v := range m {
		if k == "attributes" {
			if s, ok := v.(string); ok && s != "" {
				if parsed := parseVecMapAttributesString(s); parsed != nil {
					m[k] = parsed
				}
			}
		} else if inner, ok := v.(map[string]interface{}); ok {
			normalizeAttributesRecursive(inner)
		} else if arr, ok := v.([]interface{}); ok {
			for _, e := range arr {
				if em, ok := e.(map[string]interface{}); ok {
					normalizeAttributesRecursive(em)
				}
			}
		}
	}
}

// parseVecMapAttributesString parses JSON RPC VecMap debug string to gRPC structure
func parseVecMapAttributesString(s string) map[string]interface{} {
	keys := vecMapKeyRe.FindAllStringSubmatch(s, -1)
	values := vecMapValueRe.FindAllStringSubmatch(s, -1)
	if len(keys) == 0 || len(keys) != len(values) {
		return nil
	}
	contents := make([]map[string]string, len(keys))
	for i := range keys {
		keyStr := strings.TrimSpace(keys[i][1])
		valStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(values[i][1]), ","))
		contents[i] = map[string]string{"key": keyStr, "value": valStr}
	}
	return map[string]interface{}{"contents": contents}
}

// fixDuplicatedLinkPrefix fixes duplicated URL prefix in link
// e.g. "https://popkins.com/explorer/https://popkins.com/explorer/3029" -> "https://popkins.com/explorer/3029"
func fixDuplicatedLinkPrefix(link string) string {
	// Find second occurrence of https://, take substring from that position
	first := strings.Index(link, "https://")
	if first < 0 {
		return link
	}
	second := strings.Index(link[first+8:], "https://")
	if second < 0 {
		return link
	}
	suffix := link[first+8+second:]
	// Use suffix if it's a valid URL and there's obvious duplication
	if strings.HasPrefix(suffix, "http") && len(suffix) < len(link) {
		return suffix
	}
	return link
}

// normalizeCoinType expands short-format address (e.g. 0x2) to 64-char hex to match gRPC output
func normalizeCoinType(s string) string {
	if s == "" {
		return s
	}
	return coinTypeRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := coinTypeRe.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		hexPart, suffix := sub[1], sub[2]
		if len(hexPart) >= 64 {
			return m
		}
		return "0x" + strings.Repeat("0", 64-len(hexPart)) + hexPart + suffix
	})
}
