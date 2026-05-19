package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Schema structs
type OutSystemsSchema struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Entities    []Entity `json:"entities"`
	Actions     []Action `json:"actions"`
}

type Entity struct {
	Name       string      `json:"name"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Action struct {
	Name    string      `json:"name"`
	Inputs  []Attribute `json:"inputs"`
	Outputs []Attribute `json:"outputs"`
}

// mapOSTypeToCSharp converts OutSystems data types to C# native types.
func mapOSTypeToCSharp(osType string) string {
	switch osType {
	case "LongInteger":
		return "long"
	case "Integer":
		return "int"
	case "Text":
		return "string"
	case "Decimal":
		return "decimal"
	case "Boolean":
		return "bool"
	case "DateTime", "Date", "Time":
		return "DateTime"
	case "BinaryData":
		return "byte[]"
	default:
		return "object"
	}
}

// mapOSTypeToDefault converts OutSystems data types to C# default value strings.
func mapOSTypeToDefault(osType string) string {
	switch osType {
	case "LongInteger", "Integer":
		return "0"
	case "Text":
		return "\"\""
	case "Decimal":
		return "0m"
	case "Boolean":
		return "false"
	case "DateTime", "Date", "Time":
		return "new DateTime(1900, 1, 1, 0, 0, 0)"
	case "BinaryData":
		return "new byte[0]"
	default:
		return "null"
	}
}

// ScaffoldCSharp reads the JSON schema and writes a boilerplate C# file for an OutSystems Extension.
func ScaffoldCSharp(jsonPayload, workspaceRoot string) error {
	var schema OutSystemsSchema
	if err := json.Unmarshal([]byte(jsonPayload), &schema); err != nil {
		return fmt.Errorf("failed to parse schema for C# scaffolding: %w", err)
	}

	if schema.Name == "" {
		return fmt.Errorf("schema missing extension name")
	}

	logicDir := filepath.Join(workspaceRoot, "logic")
	if err := os.MkdirAll(logicDir, 0755); err != nil {
		return fmt.Errorf("failed to create logic directory: %w", err)
	}

	csPath := filepath.Join(logicDir, fmt.Sprintf("%s.cs", schema.Name))

	var sb strings.Builder

	sb.WriteString("using System;\n")
	sb.WriteString("using System.Collections.Generic;\n")
	sb.WriteString("using OutSystems.HubEdition.RuntimePlatform;\n")
	sb.WriteString("using OutSystems.HubEdition.RuntimePlatform.Db;\n\n")

	sb.WriteString(fmt.Sprintf("namespace OutSystems.Nss%s {\n", schema.Name))
	sb.WriteString(fmt.Sprintf("\tpublic class Css%s : Iss%s {\n\n", schema.Name, schema.Name))

	for _, action := range schema.Actions {
		sb.WriteString(fmt.Sprintf("\t\t/// <summary>\n\t\t/// %s\n\t\t/// </summary>\n", action.Name))
		
		var args []string
		
		// Input parameters
		for _, in := range action.Inputs {
			csharpType := mapOSTypeToCSharp(in.Type)
			args = append(args, fmt.Sprintf("%s ss%s", csharpType, in.Name))
		}
		
		// Output parameters
		for _, out := range action.Outputs {
			csharpType := mapOSTypeToCSharp(out.Type)
			args = append(args, fmt.Sprintf("out %s ss%s", csharpType, out.Name))
		}
		
		sb.WriteString(fmt.Sprintf("\t\tpublic void Mss%s(%s) {\n", action.Name, strings.Join(args, ", ")))
		
		sb.WriteString("\t\t\t// TODO: Write implementation for this action\n\n")

		// Initialize output parameters to prevent compilation errors
		for _, out := range action.Outputs {
			defVal := mapOSTypeToDefault(out.Type)
			sb.WriteString(fmt.Sprintf("\t\t\tss%s = %s;\n", out.Name, defVal))
		}

		sb.WriteString("\t\t}\n\n")
	}

	sb.WriteString("\t}\n") // End Class
	sb.WriteString("}\n")   // End Namespace

	if err := os.WriteFile(csPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write C# file: %w", err)
	}

	fmt.Printf("  ✅ Scaffolding complete → %s\n", csPath)
	return nil
}
