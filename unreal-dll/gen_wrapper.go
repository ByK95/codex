package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type Func struct {
	ReturnType string
	Name       string
	Params     string
	Category   string
	ParamNames []string
	BpReturnType string
	BpParams     string
	BpParamNames []string
}

type TemplateData struct {
	Filename string
	DLLName  string
	Funcs    []*Func
}

const wrapperHeaderTemplate = `
// {{ .Filename }}.h
#pragma once

#include "Kismet/BlueprintFunctionLibrary.h"
#include "{{ .Filename }}.generated.h"

UCLASS()
class ORBITSURVIVORS_API U{{ .Filename }} : public UBlueprintFunctionLibrary
{
    GENERATED_BODY()

public:
	{{- range .Funcs }}
    UFUNCTION(BlueprintCallable, Category = "{{ .Category }}")
    static {{ .BpReturnType }} {{ .Name }}({{ .BpParams }});
	{{- end }}

private:
    static bool LoadDLL();
    static void UnloadDLL();
};
`

const wrapperCPPTemplate = `
#include "{{ .Filename }}.h"
#include "HAL/PlatformProcess.h"
#include "Misc/Paths.h"
#include "Engine/Engine.h"

namespace
{
    void* DLLHandle = nullptr;
    bool bDLLInitialized = false;
{{- range .Funcs }}
    typedef {{ .ReturnType }} (*{{ .Name }}Func)({{ .Params }});
    {{ .Name }}Func p{{ .Name }} = nullptr;
{{- end }}
}

bool U{{ .Filename }}::LoadDLL()
{
    if (bDLLInitialized) return DLLHandle != nullptr;
    
    bDLLInitialized = true;
    
    FString DLLPath = FPaths::Combine(FPaths::ProjectDir(), TEXT("Binaries/Win64/{{ .DLLName }}"));
    if (!FPlatformFileManager::Get().GetPlatformFile().FileExists(*DLLPath))
    {
        UE_LOG(LogTemp, Error, TEXT("DLL not found at %s"), *DLLPath);
        return false;
    }

    DLLHandle = FPlatformProcess::GetDllHandle(*DLLPath);
    if (!DLLHandle)
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to load DLL: %s"), *DLLPath);
        return false;
    }

{{- range .Funcs }}
    p{{ .Name }} = ({{ .Name }}Func)FPlatformProcess::GetDllExport(DLLHandle, TEXT("{{ .Name }}"));
    if (!p{{ .Name }})
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to bind function: {{ .Name }}"));
    }
{{- end }}

    // Check if all functions loaded successfully
    bool bAllFunctionsLoaded = true;
{{- range .Funcs }}
    bAllFunctionsLoaded &= (p{{ .Name }} != nullptr);
{{- end }}

    if (!bAllFunctionsLoaded)
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to bind one or more DLL functions"));
        UnloadDLL();
        return false;
    }

    UE_LOG(LogTemp, Log, TEXT("{{ .DLLName }} loaded successfully"));
    return true;
}

void U{{ .Filename }}::UnloadDLL()
{
    if (DLLHandle)
    {
        FPlatformProcess::FreeDllHandle(DLLHandle);
        DLLHandle = nullptr;
    }
{{- range .Funcs }}
    p{{ .Name }} = nullptr;
{{- end }}
    bDLLInitialized = false;
}

{{- $filename := .Filename -}}
{{- range .Funcs }}

{{ .BpReturnType }} U{{ $filename }}::{{ .Name }}({{ .BpParams }})
{
    if (!LoadDLL())
    {
{{- if eq .BpReturnType "int" }}
        return -1;  // Return error value for int
{{- else if eq .BpReturnType "bool" }}
        return false;
{{- else if eq .BpReturnType "FString" }}
        return FString();
{{- else if eq .BpReturnType "void" }}
        return;
{{- else }}
        return {{ getDefaultValue .BpReturnType }};
{{- end }}
    }
	{{- range .BpParamNames }}
	FTCHARToUTF8 {{ . }}Utf8(*{{ . }});
	{{- end }}

    {{ if eq .BpReturnType "FString"}}
	char* cResult = p{{ .Name }}({{ join .ParamNames ", " }});
	FString Result = UTF8_TO_TCHAR(cResult);
	pMetrics_FreeCString(cResult);
	return Result;
    {{ else if ne .BpReturnType "void" }}return p{{ .Name }}({{ join .ParamNames ", " }});
	{{ else }}
	p{{ .Name }}({{ join .ParamNames ", " }});
	{{ end }}
}
{{- end }}
`

// Map C types to Unreal/C++ types
func mapCTypeToUnreal(cType string) string {
	cType = strings.ReplaceAll(cType, "long long int", "int64")
	switch cType {
	case "_Bool":
		return "bool"
	case "GoInt":
		return "int64"
	case "GoInt32":
		return "int32"
	case "GoUint32":
		return "uint32"
	case "GoInt64":
		return "int64"
	case "GoUint64":
		return "uint64"
	case "GoFloat32":
		return "float"
	case "GoFloat64":
		return "double"
	default:
		return cType
	}
}

func getFunctionCategory(fName string) string{
	if strings.Contains(strings.ToLower(fName), "inventory"){
		return "Inventory"
	}

	if strings.Contains(strings.ToLower(fName), "equipment"){
		return "Equipment"
	}

	if strings.Contains(strings.ToLower(fName), "metrics"){
		return "Metrics"
	}

	if strings.Contains(strings.ToLower(fName), "threat"){
		return "Threat"
	}

	if strings.Contains(strings.ToLower(fName), "store"){
		return "Store"
	}

	if strings.Contains(strings.ToLower(fName), "zoneconfig"){
		return "Zoneconfig"
	}

	return ""
}

func parseHeaderLine(line string) (Func, bool) {
	// Updated regex to handle the actual generated header format
	re := regexp.MustCompile(`extern __declspec\(dllexport\)\s+([^\s]+)\s+([^\s(]+)\(([^)]*)\);`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return Func{}, false
	}
	
	returnType := mapCTypeToUnreal(matches[1])
	name := matches[2]
	paramsRaw := strings.TrimSpace(matches[3])

	params := []string{}
	paramNames := []string{}

	if paramsRaw != "" && paramsRaw != "void" {
		parts := strings.Split(paramsRaw, ",")
		for i, p := range parts {
			p = strings.TrimSpace(p)
			
			// Map C types to Unreal types
			unrealType := mapCTypeToUnreal(p)
			
			// Generate parameter name if not present
			paramName := ""
			if strings.Contains(unrealType, " ") {
				// Type and name are present
				typeAndName := strings.Fields(unrealType)
				if len(typeAndName) >= 2 {
					paramName = typeAndName[len(typeAndName)-1]
					unrealType = strings.Join(typeAndName[:len(typeAndName)-1], " ")
				}
			}
			
			if paramName == "" {
				// Generate a parameter name
				paramName = generateParamName(unrealType, i)
			}
			
			params = append(params, unrealType+" "+paramName)
			paramNames = append(paramNames, paramName)
		}
	}

	return Func{
		ReturnType: returnType,
		Name:       name,
		Params:     strings.Join(params, ", "),
		ParamNames: paramNames,
		Category: 	getFunctionCategory(name),
		BpReturnType: returnType,
		BpParams:   strings.Join(params, ", "),
	}, true
}

func generateParamName(paramType string, index int) string {
	switch paramType {
	case "int", "int32", "int64":
		return "Value" + string(rune('A'+index))
	case "bool":
		return "bFlag" + string(rune('A'+index))
	case "float", "double":
		return "FloatValue" + string(rune('A'+index))
	default:
		return "Param" + string(rune('A'+index))
	}
}

func getDefaultValue(returnType string) string {
	switch returnType {
	case "int", "int32", "int64", "uint32", "uint64":
		return "0"
	case "bool":
		return "false"
	case "float", "double":
		return "0.0f"
	default:
		return "0"
	}
}

func extractFStringVarNames(signature string) []string {
	// Match parameters exactly like: const FString& VarName
	re := regexp.MustCompile(`const FString&\s+(\w+)`)

	matches := re.FindAllStringSubmatch(signature, -1)
	var names []string
	for _, m := range matches {
		if len(m) > 1 {
			names = append(names, m[1])
		}
	}
	return names
}

// markStringFunctions sets HasCharPtr of Functions unreal doesn't allow direct operations
// on char* so for bluprint functions need to covert it FString explicitly
func markStringFunctions(fns []*Func) {
	matcher := "char*"
	for _, fn := range fns {
		if fn.ReturnType == matcher{
			fn.BpReturnType = "FString"
		}
		if strings.Contains(fn.Params, matcher) {
			fn.BpParams = strings.ReplaceAll(fn.Params, matcher, "const FString&")
			fn.BpParamNames = extractFStringVarNames(fn.BpParams)
			
			for i, p := range fn.ParamNames {
				for _, bpName := range fn.BpParamNames {
					if p == bpName {
						fn.ParamNames[i] = fmt.Sprintf("(char*)%sUtf8.Get()", bpName)
					}
				}
			}
		}
		// b, err := json.MarshalIndent(fn, "", "  ")
		// if err != nil {
        // 	fmt.Println("error:", err)
    	// }
		// fmt.Println(string(b))
	}
}


func main() {
	filePath := "./codex.h"
	destinationPath := "E:/Unreal Projects/OrbitSurvivors/Source/OrbitSurvivors"
	destinationPublicPath := destinationPath + "/Public/"
	destinationPrivatePath := destinationPath + "/Private/"

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	funcs := []*Func{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fn, ok := parseHeaderLine(line)
		if ok {
			funcs = append(funcs, &fn)
		}
	}

	markStringFunctions(funcs)

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Create template with helper functions
	tmpl := template.Must(template.New("wrapper").Funcs(template.FuncMap{
		"join": func(list []string, sep string) string {
			return strings.Join(list, sep)
		},
		"getDefaultValue": getDefaultValue,
	}).Parse(wrapperCPPTemplate))

	tmpl2 := template.Must(template.New("wrapper").Funcs(template.FuncMap{
		"join": func(list []string, sep string) string {
			return strings.Join(list, sep)
		},
	}).Parse(wrapperHeaderTemplate))

	data := TemplateData{
		Filename: "CodexDLLBPLibrary",
		DLLName:  "codex.dll",
		Funcs:    funcs,
	}

	headerFile, err := os.Create(destinationPublicPath + "CodexDLLBPLibrary.h")
	if err != nil {
		panic(err)
	}
	defer headerFile.Close()

	cppFile, err := os.Create(destinationPrivatePath + "CodexDLLBPLibrary.cpp")
	if err != nil {
		panic(err)
	}
	defer cppFile.Close()

	err = tmpl.Execute(cppFile, data)
	if err != nil {
		panic(err)
	}

	err = tmpl2.Execute(headerFile, data)
	if err != nil {
		panic(err)
	}

	log.Printf("Generated wrapper for %d functions", len(funcs))
}