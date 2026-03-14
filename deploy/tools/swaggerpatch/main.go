package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var methods = map[string]struct{}{
	"get":     {},
	"post":    {},
	"put":     {},
	"delete":  {},
	"patch":   {},
	"head":    {},
	"options": {},
}

func main() {
	file := flag.String("file", "gateway/swagger.json", "swagger json file path")
	flag.Parse()

	if err := patchSwagger(*file); err != nil {
		fmt.Fprintf(os.Stderr, "patch swagger failed: %v\n", err)
		os.Exit(1)
	}
}

func patchSwagger(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var doc map[string]any
	if err := json.Unmarshal(content, &doc); err != nil {
		return err
	}

	doc["tags"] = []map[string]string{
		{"name": "user", "description": "用户"},
		{"name": "activity", "description": "活动"},
		{"name": "notification", "description": "通知"},
		{"name": "im", "description": "私聊"},
	}
	doc["securityDefinitions"] = map[string]any{
		"BearerAuth": map[string]any{
			"type":        "apiKey",
			"name":        "Authorization",
			"in":          "header",
			"description": "JWT Bearer token. Example: Authorization: Bearer <token>",
		},
	}
	delete(doc, "schemes")

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return fmt.Errorf("paths field missing or invalid")
	}

	for route, rawOps := range paths {
		tag := routeTag(route)
		if tag == "" {
			continue
		}

		ops, ok := rawOps.(map[string]any)
		if !ok {
			continue
		}

		for method, rawOp := range ops {
			if _, ok := methods[strings.ToLower(method)]; !ok {
				continue
			}

			op, ok := rawOp.(map[string]any)
			if !ok {
				continue
			}
			op["tags"] = []string{tag}
			delete(op, "schemes")

			if routeRequiresAuth(route) {
				op["security"] = []map[string][]string{
					{"BearerAuth": {}},
				}
			} else {
				delete(op, "security")
			}
		}
	}

	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')

	return os.WriteFile(filepath.Clean(path), encoded, 0o644)
}

func routeTag(route string) string {
	switch {
	case strings.HasPrefix(route, "/user/"):
		return "user"
	case strings.HasPrefix(route, "/activity/"):
		return "activity"
	case strings.HasPrefix(route, "/notification/"):
		return "notification"
	case strings.HasPrefix(route, "/im/"):
		return "im"
	default:
		return ""
	}
}

func routeRequiresAuth(route string) bool {
	switch {
	case route == "/user/login", route == "/user/register":
		return false
	case strings.HasPrefix(route, "/activity/public/"):
		return false
	default:
		return true
	}
}
