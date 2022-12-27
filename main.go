package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/antikuz/kmsort/internal/models"
	"gopkg.in/yaml.v2"
)

var (
	filePath = flag.String("f", "", "Path to manifests")

	files []file
)

func organizeListObjects(obj interface{}, fieldName string) interface{} {
	var result []interface{}
	objList, ok := obj.([]interface{})
	if !ok {
		log.Fatalf("cannot convert %+v to []interface {}", obj)
	}

	for _, obj := range objList {
		if objMapInterface, ok := obj.(map[interface{}]interface{}); ok {
			objMap := convertToMapString(objMapInterface)
			organizedFields := organizeFields(orderMap[fieldName], fieldName, objMap)
			result = append(result, organizedFields)
		} else {
			result = append(result, obj)
		}
	}

	return result
}

func organizeMetadata(manifestObj map[string]interface{}) yaml.MapItem {
	return yaml.MapItem{
		Key:   "metadata",
		Value: organizeFields(orderMap["metadata"], "metadata", manifestObj),
	}
}

func organizeSpec(m models.Manifest) yaml.MapItem {
	for fieldName := range m.Spec {
		if contains(ignoreSpecValues[m.Kind], fieldName) {
			delete(m.Spec, fieldName)
		}
	}

	return yaml.MapItem{
		Key:   "spec",
		Value: organizeFields(orderMap[m.Kind], "spec", m.Spec),
	}
}

func organizeTemplate(templateObj interface{}) interface{} {
	template, ok := templateObj.(map[interface{}]interface{})
	if !ok {
		log.Fatal("cannot convert template to map[string]interface{}")
	}

	metadata := convertToMapString(template["metadata"])
	spec := convertToMapString(template["spec"])

	templateManifest := models.Manifest{
		Kind:     "Template",
		Metadata: metadata,
		Spec:     spec,
	}

	templateOrderedMap := yaml.MapSlice{
		organizeMetadata(templateManifest.Metadata),
		organizeSpec(templateManifest),
	}

	return templateOrderedMap
}

func organizeMapObject(fieldNames []string, objName string, manifestObjFields interface{}) interface{} {
	var result yaml.MapSlice

	objMap := convertToMapString(manifestObjFields)

	switch objName {
	case "template":
		return organizeTemplate(manifestObjFields)
	}

	// organize preordered fields
	for _, fieldName := range fieldNames {
		if fieldObj, ok := objMap[fieldName]; ok {
			organizedFields := organizeFields(orderMap[fieldName], fieldName, fieldObj)
			result = append(result, convertToMapItem(fieldName, organizedFields))
			delete(objMap, fieldName)
		}
	}

	// organize all remaining fields in alphabetical order
	fieldNames = []string{}
    for fieldName := range objMap {
        fieldNames = append(fieldNames, fieldName)
    }
    sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
		fieldObj := objMap[fieldName]
		if _, ok := orderMap[fieldName]; ok {
			switch fieldObj.(type) {
			case []interface{}:
				organizedFields := organizeListObjects(fieldObj, fieldName)
				result = append(result, convertToMapItem(fieldName, organizedFields))
			default:
				subObjMap := convertToMapString(fieldObj.(map[interface{}]interface{}))
				organizedFields := organizeFields(orderMap[fieldName], fieldName, subObjMap)
				result = append(result, convertToMapItem(fieldName, organizedFields))
			}
		} else if value, ok := fieldObj.(map[interface{}]interface{}); ok {
			if len(value) != 0 || fieldName == "emptyDir" {
				subObjMap := convertToMapString(value)
				organizedFields := organizeFields(orderMap[fieldName], fieldName, subObjMap)
				result = append(result, convertToMapItem(fieldName, organizedFields))
			}
		} else if !contains(fieldsPopulatedByTheSystem, fieldName) {
			if !fieldValueIdenticalToDefaultValue(fieldName, fieldObj) {
				result = append(result, convertToMapItem(fieldName, fieldObj))
			}
		}
		delete(objMap, fieldName)
	}

	return result
}

func organizeFields(fieldNames []string, manifestObjName string, manifestObjFields interface{}) interface{} {
	var result yaml.MapSlice
	var organizedFields interface{}

	switch manifestObjFields.(type) {
	case map[interface{}]interface{}:
		organizedFields = organizeMapObject(fieldNames, manifestObjName, manifestObjFields)
	case map[string]interface{}:
		organizedFields = organizeMapObject(fieldNames, manifestObjName, manifestObjFields)
	case []interface{}:
		organizedFields = organizeListObjects(manifestObjFields, manifestObjName)
	default:
		organizedFields = manifestObjFields
	}

	if mapItem, ok := organizedFields.(yaml.MapItem); ok {
		result = append(result, mapItem)
		return result
	}

	return organizedFields
}

type file struct {
	name string
	dir  string
}

func sortManifest(manifestByte []byte) ([]byte, error) {
	manifest := models.Manifest{}
	err := yaml.Unmarshal(manifestByte, &manifest)
	if err != nil {
		return nil, err
	}
	manifestOrderedMap := yaml.MapSlice{
		convertToMapItem("apiVersion", manifest.ApiVersion),
		convertToMapItem("kind", manifest.Kind),
		organizeMetadata(manifest.Metadata),
	}

	if manifest.Kind == "ConfigMap" || manifest.Kind == "Secret" {
		if manifest.Data != nil {
			manifestOrderedMap = append(manifestOrderedMap, convertToMapItem("data", manifest.Data))
		} else if manifest.StringData != nil  {
			manifestOrderedMap = append(manifestOrderedMap, convertToMapItem("stringData", manifest.StringData))
		}
	} else {
		manifestOrderedMap = append(manifestOrderedMap, organizeSpec(manifest))
	}

	manifestByte, err = yaml.Marshal(&manifestOrderedMap)
	if err != nil {
		return nil, err
	}

	return manifestByte, nil
}


func filesProcessing() {
	fileInfo, err := os.Stat(*filePath)
	if err != nil {
		log.Fatalf("Cannot open path %s, cause err: %v", *filePath, err)
	}

	if fileInfo.IsDir() {
		err = filepath.Walk(*filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				dir, name := filepath.Split(path)
				files = append(files, file{
					name: name,
					dir:  dir,
				})
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	} else {
		files = append(files, file{
			name: *filePath,
			dir:  ".",
		})
	}

	for _, f := range files {
		stream, err := ioutil.ReadFile(filepath.Join(f.dir, f.name))
		if err != nil {
			log.Fatal(err)
		}

		manifestsByte := bytes.Split(stream, []byte("---"))
		result := [][]byte{}
		for _, manifestByte := range manifestsByte {
			sortedManifest, err := sortManifest(manifestByte)
			if err != nil {
				log.Fatal(err)
			}

			result = append(result, sortedManifest)
		}

		manifestByte := bytes.Join(result, []byte("---\n"))
		err = os.MkdirAll(filepath.Join("new", f.dir), 0644)
		if err != nil {
			log.Fatal(err)
		}
		
		err = ioutil.WriteFile(filepath.Join("new", f.dir, f.name), manifestByte, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	flag.Parse()
	if *filePath == "" {
		startWebserver()
	} else {
		filesProcessing()
	}
}
