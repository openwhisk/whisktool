/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parsers

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"strings"

	"encoding/base64"
	"encoding/json"

	"github.com/apache/incubator-openwhisk-client-go/whisk"
	"github.com/apache/incubator-openwhisk-wskdeploy/utils"
	"gopkg.in/yaml.v2"
	"fmt"
)

// Read existing manifest file or create new if none exists
func ReadOrCreateManifest() *ManifestYAML {
	maniyaml := ManifestYAML{}

	if _, err := os.Stat("manifest.yaml"); err == nil {
		dat, _ := ioutil.ReadFile("manifest.yaml")
		err := NewYAMLParser().Unmarshal(dat, &maniyaml)
		utils.Check(err)
	}
	return &maniyaml
}

// Serialize manifest to local file
func Write(manifest *ManifestYAML, filename string) {
	output, err := NewYAMLParser().Marshal(manifest)
	utils.Check(err)

	f, err := os.Create(filename)
	utils.Check(err)
	defer f.Close()

	f.Write(output)
}

func (dm *YAMLParser) Unmarshal(input []byte, manifest *ManifestYAML) error {
	err := yaml.Unmarshal(input, manifest)
	if err != nil {
		log.Printf("error happened during unmarshal :%v", err)
		return err
	}
	return nil
}

func (dm *YAMLParser) Marshal(manifest *ManifestYAML) (output []byte, err error) {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		log.Printf("err happened during marshal :%v", err)
		return nil, err
	}
	return data, nil
}

func (dm *YAMLParser) ParseManifest(mani string) *ManifestYAML {
	mm := NewYAMLParser()
	maniyaml := ManifestYAML{}

	content, err := utils.Read(mani)
	utils.Check(err)

	err = mm.Unmarshal(content, &maniyaml)
	utils.Check(err)
	maniyaml.Filepath = mani
	return &maniyaml
}

func (dm *YAMLParser) ComposeDependencies(mani *ManifestYAML, projectPath string) (map[string]utils.DependencyRecord, error) {

	var errorParser error
	depMap := make(map[string]utils.DependencyRecord)
	for key, dependency := range mani.Package.Dependencies {
		version := dependency.Version
		if version == "" {
			version = "master"
		}

		location := dependency.Location

		isBinding := false
		if utils.LocationIsBinding(location) {

			if !strings.HasPrefix(location, "/") {
				location = "/" + dependency.Location
			}

			isBinding = true
		} else if utils.LocationIsGithub(location) {

			if !strings.HasPrefix(location, "https://") && !strings.HasPrefix(location, "http://") {
				location = "https://" + dependency.Location
			}

			isBinding = false
		} else {
			return nil, errors.New("Dependency type is unknown.  wskdeploy only supports /whisk.system bindings or github.com packages.")
		}

		keyValArrParams := make(whisk.KeyValueArr, 0)
		for name, param := range dependency.Inputs {
			var keyVal whisk.KeyValue
			keyVal.Key = name

			keyVal.Value, errorParser = ResolveParameter(&param)

			if errorParser != nil {
				return nil, errorParser
			}

			if keyVal.Value != nil {
				keyValArrParams = append(keyValArrParams, keyVal)
			}
		}

		keyValArrAnot := make(whisk.KeyValueArr, 0)
		for name, value := range dependency.Annotations {
			var keyVal whisk.KeyValue
			keyVal.Key = name
			keyVal.Value = utils.GetEnvVar(value)

			keyValArrAnot = append(keyValArrAnot, keyVal)
		}

		packDir := path.Join(projectPath, "Packages")
		depMap[key] = utils.DependencyRecord{packDir, mani.Package.Packagename, location, version, keyValArrParams, keyValArrAnot, isBinding}
	}

	return depMap, nil
}

// Is we consider multi pacakge in one yaml?
func (dm *YAMLParser) ComposePackage(mani *ManifestYAML) (*whisk.Package, error) {
	var errorParser error

	//mani := dm.ParseManifest(manipath)
	pag := &whisk.Package{}
	pag.Name = mani.Package.Packagename
	//The namespace for this package is absent, so we use default guest here.
	pag.Namespace = mani.Package.Namespace
	pub := false
	pag.Publish = &pub

	keyValArr := make(whisk.KeyValueArr, 0)
	for name, param := range mani.Package.Inputs {
		var keyVal whisk.KeyValue
		keyVal.Key = name

		keyVal.Value, errorParser = ResolveParameter(&param)

		if errorParser != nil {
			return nil, errorParser
		}

		if keyVal.Value != nil {
			keyValArr = append(keyValArr, keyVal)
		}
	}

	if len(keyValArr) > 0 {
		pag.Parameters = keyValArr
	}
	return pag, nil
}

func (dm *YAMLParser) ComposeSequences(namespace string, mani *ManifestYAML) ([]utils.ActionRecord, error) {
	var s1 []utils.ActionRecord = make([]utils.ActionRecord, 0)
	for key, sequence := range mani.Package.Sequences {
		wskaction := new(whisk.Action)
		wskaction.Exec = new(whisk.Exec)
		wskaction.Exec.Kind = "sequence"
		actionList := strings.Split(sequence.Actions, ",")

		var components []string
		for _, a := range actionList {

			act := strings.TrimSpace(a)

			if !strings.ContainsRune(act, '/') && !strings.HasPrefix(act, mani.Package.Packagename+"/") {
				act = path.Join(mani.Package.Packagename, act)
			}
			components = append(components, path.Join("/"+namespace, act))
		}

		wskaction.Exec.Components = components
		wskaction.Name = key
		pub := false
		wskaction.Publish = &pub
		wskaction.Namespace = namespace

		keyValArr := make(whisk.KeyValueArr, 0)
		for name, value := range sequence.Annotations {
			var keyVal whisk.KeyValue
			keyVal.Key = name
			keyVal.Value = utils.GetEnvVar(value)

			keyValArr = append(keyValArr, keyVal)
		}

		if len(keyValArr) > 0 {
			wskaction.Annotations = keyValArr
		}

		record := utils.ActionRecord{wskaction, mani.Package.Packagename, key}
		s1 = append(s1, record)
	}
	return s1, nil
}

func (dm *YAMLParser) ComposeActions(mani *ManifestYAML, manipath string) (ar []utils.ActionRecord, aub []*utils.ActionExposedURLBinding, err error) {

	var errorParser error
	var s1 []utils.ActionRecord = make([]utils.ActionRecord, 0)
	var au []*utils.ActionExposedURLBinding = make([]*utils.ActionExposedURLBinding, 0)

	for key, action := range mani.Package.Actions {
		splitmanipath := strings.Split(manipath, string(os.PathSeparator))

		wskaction := new(whisk.Action)
		//bind action, and exposed URL
		aubinding := new(utils.ActionExposedURLBinding)
		aubinding.ActionName = key
		aubinding.ExposedUrl = action.ExposedUrl

		wskaction.Exec = new(whisk.Exec)
		if action.Location != "" {
			filePath := strings.TrimRight(manipath, splitmanipath[len(splitmanipath)-1]) + action.Location

			if utils.IsDirectory(filePath) {
				zipName := filePath + ".zip"
				err = utils.NewZipWritter(filePath, zipName).Zip()
				defer os.Remove(zipName)
				utils.Check(err)
				// To do: support docker and main entry as did by go cli?
				wskaction.Exec, err = utils.GetExec(zipName, action.Runtime, false, "")
			} else {
				ext := path.Ext(filePath)
				kind := "nodejs:default"

				switch ext {
				case ".swift":
					kind = "swift:default"
				case ".js":
					kind = "nodejs:default"
				case ".py":
					kind = "python"
				}

				wskaction.Exec.Kind = kind

				action.Location = filePath
				dat, err := utils.Read(filePath)
				utils.Check(err)
				code := string(dat)
				if ext == ".zip" || ext == ".jar" {
					code = base64.StdEncoding.EncodeToString([]byte(dat))
				}
				wskaction.Exec.Code = &code
			}

		}

		if action.Runtime != "" {
			wskaction.Exec.Kind = action.Runtime
		}

		// we can specify the name of the action entry point using main
		if action.Main != "" {
			wskaction.Exec.Main = action.Main
		}

		keyValArr := make(whisk.KeyValueArr, 0)
		for name, param := range action.Inputs {
			var keyVal whisk.KeyValue
			keyVal.Key = name
			println("NAME: " + name)

			keyVal.Value, errorParser = ResolveParameter(&param)

			if errorParser != nil {
				return nil, nil, errorParser
			}

			if keyVal.Value != nil {
				keyValArr = append(keyValArr, keyVal)
			}
		}

		if len(keyValArr) > 0 {
			wskaction.Parameters = keyValArr
		}

		keyValArr = make(whisk.KeyValueArr, 0)
		for name, value := range action.Annotations {
			var keyVal whisk.KeyValue
			keyVal.Key = name
			keyVal.Value = utils.GetEnvVar(value)

			keyValArr = append(keyValArr, keyVal)
		}

		// only set the webaction when the annotations are not empty.
		if len(keyValArr) > 0 && action.Webexport == "true" {
			//wskaction.Annotations = keyValArr
			wskaction.Annotations, err = utils.WebAction("yes", keyValArr, action.Name, false)
			utils.Check(err)
		}

		wskaction.Name = key
		pub := false
		wskaction.Publish = &pub

		record := utils.ActionRecord{wskaction, mani.Package.Packagename, action.Location}
		s1 = append(s1, record)

		//only append when the fields are exists
		if aubinding.ActionName != "" && aubinding.ExposedUrl != "" {
			au = append(au, aubinding)
		}

	}

	return s1, au, nil

}

func (dm *YAMLParser) ComposeTriggers(manifest *ManifestYAML) ([]*whisk.Trigger, error) {
	var errorParser error
	var t1 []*whisk.Trigger = make([]*whisk.Trigger, 0)
	pkg := manifest.Package
	for _, trigger := range pkg.GetTriggerList() {
		wsktrigger := new(whisk.Trigger)
		wsktrigger.Name = trigger.Name
		wsktrigger.Namespace = trigger.Namespace
		pub := false
		wsktrigger.Publish = &pub

		keyValArr := make(whisk.KeyValueArr, 0)
		if trigger.Source != "" {
			var keyVal whisk.KeyValue

			keyVal.Key = "feed"
			keyVal.Value = trigger.Source

			keyValArr = append(keyValArr, keyVal)

			wsktrigger.Annotations = keyValArr
		}

		keyValArr = make(whisk.KeyValueArr, 0)
		for name, param := range trigger.Inputs {
			var keyVal whisk.KeyValue
			keyVal.Key = name

			keyVal.Value, errorParser = ResolveParameter(&param)

			if errorParser != nil {
				return nil, errorParser
			}

			if keyVal.Value != nil {
				keyValArr = append(keyValArr, keyVal)
			}
		}

		if len(keyValArr) > 0 {
			wsktrigger.Parameters = keyValArr
		}

		t1 = append(t1, wsktrigger)
	}
	return t1, nil
}

func (dm *YAMLParser) ComposeRules(manifest *ManifestYAML) ([]*whisk.Rule, error) {

	var r1 []*whisk.Rule = make([]*whisk.Rule, 0)
	pkg := manifest.Package
	for _, rule := range pkg.GetRuleList() {
		wskrule := rule.ComposeWskRule()

		act := strings.TrimSpace(wskrule.Action.(string))

		if !strings.ContainsRune(act, '/') && !strings.HasPrefix(act, pkg.Packagename+"/") {
			act = path.Join(pkg.Packagename, act)
		}

		wskrule.Action = act

		r1 = append(r1, wskrule)
	}

	return r1, nil
}

func (action *Action) ComposeWskAction(manipath string) (*whisk.Action, error) {
	wskaction, err := utils.CreateActionFromFile(manipath, action.Location)
	utils.Check(err)
	wskaction.Name = action.Name
	wskaction.Version = action.Version
	wskaction.Namespace = action.Namespace
	return wskaction, err
}


// TODO() Support other valid Package Manifest types
// TODO() i.e., json (valid), timestamp, version, string256, string64, string16
// TODO() Support JSON schema validation for type: json
// TODO(): Support OpenAPI schema validation

var validParameterNameMap = map[string]string{
	"string": "string",
	"int": "integer",
	"float": "float",
	"bool": "boolean",
	"int8": "integer",
	"int16": "integer",
	"int32": "integer",
	"int64": "integer",
	"float32": "float",
	"float64": "float",
}


var typeDefaultValueMap = map[string]interface{} {
	"string": "",
	"integer": 0,
	"float": 0.0,
	"boolean": false,
	// @TODO() Support these types + their validation
	// timestamp
	// null
	// version
	// string256
	// string64
	// string16
	// json
	// scalar-unit
	// schema
	// object
}

func isValidParameterType(typeName string) bool {
	_, isValid := typeDefaultValueMap[typeName]
	return isValid
}

// TODO() throw errors
func getTypeDefaultValue(typeName string) interface{} {

	if val, ok := typeDefaultValueMap[typeName]; ok {
		return val
	} else {
		// TODO() throw an error "type not found"
	}
        return nil
}

func ResolveParamTypeFromValue(value interface{}) (string, error) {
        // Note: string is the default type if not specified.
	var paramType string = "string"
	var err error = nil

	if value != nil {
		actualType := reflect.TypeOf(value).Kind().String()

		// See if the actual type of the value is valid
		if normalizedTypeName, found := validParameterNameMap[actualType]; found {
			// use the full spec. name
			paramType = normalizedTypeName

		} else {
			// raise an error if param is not a known type
			err = utils.NewParserErr("",-1, "Parameter value is not a known type. [" + actualType + "]")
		}
	} else {

		// TODO: The value may be supplied later, we need to support non-fatal warnings
		// raise an error if param is nil
		//err = utils.NewParserErr("",-1,"Paramter value is nil.")
	}
	return paramType, err
}

// Resolve input parameter (i.e., type, value, default)
// Note: parameter values may set later (overriddNen) by an (optional) Deployment file
func ResolveParameter(param *Parameter) (interface{}, error) {

	var errorParser error
	// default parameter value to empty string
	var value interface{} = ""

	dumpParameter("BEFORE", param)

	// Parameters can be single OR multi-line declarations which must be processed/validated differently
	if !param.multiline {
		// we have a single-line parameter declaration
		// We need to identify parameter Type here for later validation
		param.Type, errorParser = ResolveParamTypeFromValue(param.Value)

	} else {
		// we have a multi-line parameter declaration


	}

	// Make sure the parameter's value is a valid, non-empty string and startsWith '$" (dollar) sign
	value = utils.GetEnvVar(param.Value)

	typ := param.Type
	// if value is of type 'string' and its not empty <OR> if type is not 'string'
	// TODO(): need to validate type is one of the supported primitive types with unit testing
	if str, ok := value.(string); ok && (len(typ) == 0 || typ != "string") {
		var parsed interface{}
		err := json.Unmarshal([]byte(str), &parsed)
		if err == nil {
			return parsed, err
		}
	}

	dumpParameter("AFTER", param)
	fmt.Printf("EXIT: value=[%v]\n", value)

	// @TODO() Need warning message here, support for warnings (non-fatal)
	// Default to an empty string, do NOT error/terminate as Value may be provided later bu a Deployment file.
	if (value == nil) {
		value = ""
		param.Type = "string"
	}
	return value, errorParser
}

// Provide custom Parameter marshalling and unmarshalling

type ParsedParameter Parameter

func (n *Parameter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux ParsedParameter

	// Attempt to unmarshall the multi-line schema
	if err := unmarshal(&aux); err == nil {
		n.multiline = true
		n.Type = aux.Type
		n.Description = aux.Description
		n.Value = aux.Value
		n.Required = aux.Required
		n.Default = aux.Default
		n.Status = aux.Status
		n.Schema = aux.Schema
		return nil
	} else {

	}

	// If we did not find the multi-line schema, assume in-line (or single-line) schema
	var inline interface{}
	if err := unmarshal(&inline); err != nil {
		return err
	}

	n.Value = inline
	n.multiline = false
	return nil
}

func (n *Parameter) MarshalYAML() (interface{}, error) {
	if _, ok := n.Value.(string); len(n.Type) == 0 && len(n.Description) == 0 && ok {
		if !n.Required && len(n.Status) == 0 && n.Schema == nil {
			return n.Value.(string), nil
		}
	}

	return n, nil
}

func dumpParameter(sep string, param *Parameter) {

	fmt.Printf("%s: %T\n", sep, param)
	if(param!= nil) {
		fmt.Printf("\tParameter.Type: [%s]\n", param.Type)

		//var str string = param.Value.(string)
		fmt.Printf("\tParameter.Value: [%v]\n", param.Value)
		fmt.Printf("\tParameter.Default: [%v]\n", param.Default)
	}
}
