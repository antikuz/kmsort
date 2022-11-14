package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
		objMap := convertToMapString(obj.(map[interface{}]interface{}))
		organizedFields := organizeFields(orderMap[fieldName], fieldName, objMap)
		result = append(result, organizedFields)
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

	for fieldName, fieldObj := range objMap {
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
			// fmt.Printf("Key:%s, Value:%s (%T), Len:%d\n", fieldName, fieldValue, fieldValue, len(value))
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

func main() {
	flag.Parse()
	if *filePath == "" {
		flag.Usage()
		os.Exit(0)
	}
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
		// fileName := "lcm-deployment.yaml"
		stream, err := ioutil.ReadFile(filepath.Join(f.dir, f.name))
		if err != nil {
			log.Fatal(err)
		}

		manifest := models.Manifest{}
		err = yaml.Unmarshal(stream, &manifest)
		if err != nil {
			log.Fatal(err)
		}

		manifestOrderedMap := yaml.MapSlice{
			convertToMapItem("apiVersion", manifest.ApiVersion),
			convertToMapItem("kind", manifest.Kind),
			organizeMetadata(manifest.Metadata),
		}

		if manifest.Kind == "ConfigMap" || manifest.Kind == "Secret" {
			manifestOrderedMap = append(manifestOrderedMap, convertToMapItem("data", manifest.Data))
		} else {
			manifestOrderedMap = append(manifestOrderedMap, organizeSpec(manifest))
		}

		manifestByte, _ := yaml.Marshal(&manifestOrderedMap)

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