// +build unit

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
    "github.com/apache/incubator-openwhisk-wskdeploy/utils"
    "github.com/stretchr/testify/assert"
    "io/ioutil"
    "os"
    "testing"
    "fmt"
    "path/filepath"
    "reflect"
    "strconv"
    "strings"
)

// Test 1: validate manifest_parser:Unmarshal() method with a sample manifest in NodeJS
// validate that manifest_parser is able to read and parse the manifest data
func TestUnmarshalForHelloNodeJS(t *testing.T) {
    data := `
package:
  name: helloworld
  actions:
    helloNodejs:
      function: actions/hello.js
      runtime: nodejs:6`
    // set the zero value of struct YAML
    m := YAML{}
    // Unmarshal reads/parses manifest data and sets the values of YAML
    // And returns an error if parsing a manifest data fails
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    if err == nil {
        // YAML.Filepath does not get set by Parsers.Unmarshal
        // as it takes manifest YAML data as a function parameter
        // instead of file name of a manifest file, therefore there is
        // no way for Unmarshal function to set YAML.Filepath field
        // (TODO) Ideally we should change this functionality so that
        // (TODO) filepath is set to the actual path of the manifest file
        expectedResult := ""
        actualResult := m.Filepath
        assert.Equal(t, expectedResult, actualResult, "Expected filepath to be an empty" +
            " string instead its set to " + actualResult + " which is invalid value")
        // package name should be "helloworld"
        expectedResult = "helloworld"
        actualResult = m.Package.Packagename
        assert.Equal(t, expectedResult, actualResult, "Expected package name " + expectedResult + " but got " + actualResult)
        // manifest should contain only one action
        expectedResult = string(1)
        actualResult = string(len(m.Package.Actions))
        assert.Equal(t, expectedResult, actualResult, "Expected 1 but got " + actualResult)
        // get the action payload from the map of actions which is stored in
        // YAML.Package.Actions with the type of map[string]Action
        actionName := "helloNodejs"
        if action, ok := m.Package.Actions[actionName]; ok {
            // location/function of an action should be "actions/hello.js"
            expectedResult = "actions/hello.js"
            actualResult = action.Function
            assert.Equal(t, expectedResult, actualResult, "Expected action function " + expectedResult + " but got " + actualResult)
            // runtime of an action should be "nodejs:6"
            expectedResult = "nodejs:6"
            actualResult = action.Runtime
            assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)
        } else {
            t.Error("Action named " + actionName + " does not exist.")
        }
    }
}

// Test 2: validate manifest_parser:Unmarshal() method with a sample manifest in Java
// validate that manifest_parser is able to read and parse the manifest data
func TestUnmarshalForHelloJava(t *testing.T) {
    data := `
package:
  name: helloworld
  actions:
    helloJava:
      function: actions/hello.jar
      runtime: java
      main: Hello`
    m := YAML{}
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    // nothing to test if Unmarshal returns an err
    if err == nil {
        // get an action from map of actions where key is action name and
        // value is Action struct
        actionName := "helloJava"
        if action, ok := m.Package.Actions[actionName]; ok {
            // runtime of an action should be java
            expectedResult := "java"
            actualResult := action.Runtime
            assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)
            // Main field should be set to "Hello"
            expectedResult = action.Main
            actualResult = "Hello"
            assert.Equal(t, expectedResult, actualResult, "Expected action main function " + expectedResult + " but got " + actualResult)
        } else {
            t.Error("Expected action named " + actionName + " but does not exist.")
        }
    }
}

// Test 3: validate manifest_parser:Unmarshal() method with a sample manifest in Python
// validate that manifest_parser is able to read and parse the manifest data
func TestUnmarshalForHelloPython(t *testing.T) {
    data := `
package:
  name: helloworld
  actions:
    helloPython:
      function: actions/hello.py
      runtime: python`
    m := YAML{}
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    // nothing to test if Unmarshal returns an err
    if err == nil {
        // get an action from map of actions which is defined as map[string]Action{}
        actionName := "helloPython"
        if action, ok := m.Package.Actions[actionName]; ok {
            // runtime of an action should be python
            expectedResult := "python"
            actualResult := action.Runtime
            assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)
        } else {
            t.Error("Expected action named " + actionName + " but does not exist.")
        }
    }
}

// Test 4: validate manifest_parser:Unmarshal() method with a sample manifest in Swift
// validate that manifest_parser is able to read and parse the manifest data
func TestUnmarshalForHelloSwift(t *testing.T) {
    data := `
package:
  name: helloworld
  actions:
    helloSwift:
      function: actions/hello.swift
      runtime: swift`
    m := YAML{}
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    // nothing to test if Unmarshal returns an err
    if err == nil {
        // get an action from map of actions which is defined as map[string]Action{}
        actionName := "helloSwift"
        if action, ok := m.Package.Actions[actionName]; ok {
            // runtime of an action should be swift
            expectedResult := "swift"
            actualResult := action.Runtime
            assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)
        } else {
            t.Error("Expected action named " + actionName + " but does not exist.")
        }
    }
}

// Test 5: validate manifest_parser:Unmarshal() method for an action with parameters
// validate that manifest_parser is able to read and parse the manifest data, specially
// validate two input parameters and their values
func TestUnmarshalForHelloWithParams(t *testing.T) {
    var data = `
package:
   name: helloworld
   actions:
     helloWithParams:
       function: actions/hello-with-params.js
       runtime: nodejs:6
       inputs:
         name: Amy
         place: Paris`
    m := YAML{}
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    if err == nil {
        actionName := "helloWithParams"
        if action, ok := m.Package.Actions[actionName]; ok {
            expectedResult := "Amy"
            actualResult := action.Inputs["name"].Value.(string)
            assert.Equal(t, expectedResult, actualResult,
                "Expected input parameter " + expectedResult + " but got " + actualResult + "for name")
            expectedResult = "Paris"
            actualResult = action.Inputs["place"].Value.(string)
            assert.Equal(t, expectedResult, actualResult,
                "Expected input parameter " + expectedResult + " but got " + actualResult + "for place")
        }
    }
}

// Test 6: validate manifest_parser:Unmarshal() method for an invalid manifest
// manifest_parser should report an error when a package section is missing
func TestUnmarshalForMissingPackage(t *testing.T) {
    data := `
  actions:
    helloNodejs:
      function: actions/hello.js
      runtime: nodejs:6
    helloJava:
      function: actions/hello.java`
    // set the zero value of struct YAML
    m := YAML{}
    // Unmarshal reads/parses manifest data and sets the values of YAML
    // And returns an error if parsing a manifest data fails
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    assert.NotNil(t, err, "Expected some error from Unmarshal but got no error")

}

/*
 Test 7: validate manifest_parser:ParseManifest() method for multiline parameters
 manifest_parser should be able to parse all different mutliline combinations of
 inputs section including:

 case 1: value only
 param:
    value: <value>
 case 2: type only
 param:
    type: <type>
 case 3: type and value only
 param:
    type: <type>
    value: <value>
 case 4: default value
 param:
    type: <type>
    default: <default value>
*/
func TestParseManifestForMultiLineParams(t *testing.T) {
    // manifest file is located under ../tests folder
    manifestFile := "../tests/dat/manifest_validate_multiline_params.yaml"
    // read and parse manifest.yaml file
    m, _ := NewYAMLParser().ParseManifest(manifestFile)

    // validate package name should be "validate"
    packageName := "validate"
    assert.NotNil(t, m.Packages[packageName],
        "Expected package named validate but got none")

    // validate this package contains one action
    expectedActionsCount := 1
    actualActionsCount := len(m.Packages[packageName].Actions)
    assert.Equal(t, expectedActionsCount, actualActionsCount,
        "Expected " + string(expectedActionsCount) + " but got " + string(actualActionsCount))

    // here Package.Actions holds a map of map[string]Action
    // where string is the action name so in case you create two actions with
    // same name, will go unnoticed
    // also, the Action struct does not have name field set it to action name
    actionName := "validate_multiline_params"
    if action, ok := m.Packages[packageName].Actions[actionName]; ok {
        // validate location/function of an action to be "actions/dump_params.js"
        expectedResult := "actions/dump_params.js"
        actualResult := action.Function
        assert.Equal(t, expectedResult, actualResult, "Expected action function " + expectedResult + " but got " + actualResult)

        // validate runtime of an action to be "nodejs:6"
        expectedResult = "nodejs:6"
        actualResult = action.Runtime
        assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)

        // validate the number of inputs to this action
        expectedResult = strconv.FormatInt(10, 10)
        actualResult = strconv.FormatInt(int64(len(action.Inputs)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // validate inputs to this action
        for input, param := range action.Inputs {
            switch input {
            case "param_string_value_only":
                expectedResult = "foo"
                actualResult = param.Value.(string)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_int_value_only":
                expectedResult = strconv.FormatInt(123, 10)
                actualResult = strconv.FormatInt(int64(param.Value.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_float_value_only":
                expectedResult = strconv.FormatFloat(3.14, 'f', -1, 64)
                actualResult = strconv.FormatFloat(param.Value.(float64), 'f', -1, 64)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_string_type_and_value_only":
                expectedResult = "foo"
                actualResult = param.Value.(string)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
                expectedResult = "string"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_string_type_only":
                expectedResult = "string"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_integer_type_only":
                expectedResult = "integer"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_float_type_only":
                expectedResult = "float"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_string_with_default":
                expectedResult = "string"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
                expectedResult = "bar"
                actualResult = param.Default.(string)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_integer_with_default":
                expectedResult = "integer"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
                expectedResult = strconv.FormatInt(-1, 10)
                actualResult = strconv.FormatInt(int64(param.Default.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_float_with_default":
                expectedResult = "float"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
                expectedResult = strconv.FormatFloat(2.9, 'f', -1, 64)
                actualResult = strconv.FormatFloat(param.Default.(float64), 'f', -1, 64)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            }
        }

        // validate outputs
        // output payload is of type string and has a description
        //if payload, ok := action.Outputs["payload"]; ok {
        //    p := payload.(map[interface{}]interface{})
        //    expectedResult = "string"
        //    actualResult = p["type"].(string)
        //    assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
        //    expectedResult = "parameter dump"
        //    actualResult = p["description"].(string)
        //    assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
        //}
    }
}

// Test 8: validate manifest_parser:ParseManifest() method for single line parameters
// manifest_parser should be able to parse input section with different types of values
func TestParseManifestForSingleLineParams(t *testing.T) {
    // manifest file is located under ../tests folder
    manifestFile := "../tests/dat/manifest_validate_singleline_params.yaml"
    // read and parse manifest.yaml file
    m, _ := NewYAMLParser().ParseManifest(manifestFile)

    // validate package name should be "validate"
    packageName := "validate"
    assert.NotNil(t, m.Packages[packageName],
        "Expected package named "+ packageName + " but got none")

    // validate this package contains one action
    expectedActionsCount := 1
    actualActionsCount := len(m.Packages[packageName].Actions)
    assert.Equal(t, expectedActionsCount, actualActionsCount,
        "Expected " + string(expectedActionsCount) + " but got " + string(actualActionsCount))

    actionName := "validate_singleline_params"
    if action, ok := m.Packages[packageName].Actions[actionName]; ok {
        // validate location/function of an action to be "actions/dump_params.js"
        expectedResult := "actions/dump_params.js"
        actualResult := action.Function
        assert.Equal(t, expectedResult, actualResult, "Expected action function " + expectedResult + " but got " + actualResult)

        // validate runtime of an action to be "nodejs:6"
        expectedResult = "nodejs:6"
        actualResult = action.Runtime
        assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)

        // validate the number of inputs to this action
        expectedResult = strconv.FormatInt(22, 10)
        actualResult = strconv.FormatInt(int64(len(action.Inputs)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // validate Inputs to this action
        for input, param := range action.Inputs {
            switch input {
            case "param_simple_string":
                expectedResult = "foo"
                actualResult = param.Value.(string)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_integer_1":
                expectedResult = strconv.FormatInt(1, 10)
                actualResult = strconv.FormatInt(int64(param.Value.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_integer_2":
                expectedResult = strconv.FormatInt(0, 10)
                actualResult = strconv.FormatInt(int64(param.Value.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_integer_3":
                expectedResult = strconv.FormatInt(-1, 10)
                actualResult = strconv.FormatInt(int64(param.Value.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_integer_4":
                expectedResult = strconv.FormatInt(99999, 10)
                actualResult = strconv.FormatInt(int64(param.Value.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_integer_5":
                expectedResult = strconv.FormatInt(-99999, 10)
                actualResult = strconv.FormatInt(int64(param.Value.(int)), 10)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_float_1":
                expectedResult = strconv.FormatFloat(1.1, 'f', -1, 64)
                actualResult = strconv.FormatFloat(param.Value.(float64), 'f', -1, 64)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_float_2":
                expectedResult = strconv.FormatFloat(0.0, 'f', -1, 64)
                actualResult = strconv.FormatFloat(param.Value.(float64), 'f', -1, 64)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_float_3":
                expectedResult = strconv.FormatFloat(-1.1, 'f', -1, 64)
                actualResult = strconv.FormatFloat(param.Value.(float64), 'f', -1, 64)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_env_var_1":
                expectedResult = "$GOPATH"
                actualResult = param.Value.(string)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_invalid_env_var":
                expectedResult = "$DollarSignNotInEnv"
                actualResult = param.Value.(string)
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            case "param_simple_implied_empty":
                assert.Nil(t, param.Value, "Expected nil")
            case "param_simple_explicit_empty_1":
                actualResult = param.Value.(string)
                assert.Empty(t, actualResult, "Expected empty string but got " + actualResult)
            case "param_simple_explicit_empty_2":
                actualResult = param.Value.(string)
                assert.Empty(t, actualResult, "Expected empty string but got " + actualResult)
            }
        }

        // validate Outputs from this action
        for output, param := range action.Outputs {
            switch output {
            case "payload":
                expectedResult = "string"
                actualResult = param.Type
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

                expectedResult = "parameter dump"
                actualResult = param.Description
                assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
            }
        }
    }
}

// Test 9: validate manifest_parser.ComposeActions() method for implicit runtimes
// when a runtime of an action is not provided, manifest_parser determines the runtime
// based on the file extension of an action file
func TestComposeActionsForImplicitRuntimes(t *testing.T) {
    data :=
        `package:
  name: helloworld
  actions:
    helloNodejs:
      function: ../tests/src/integration/helloworld/actions/hello.js
    helloJava:
      function: ../tests/src/integration/helloworld/actions/hello.jar
      main: Hello
    helloPython:
      function: ../tests/src/integration/helloworld/actions/hello.py
    helloSwift:
      function: ../tests/src/integration/helloworld/actions/hello.swift`

    dir, _ := os.Getwd()
    tmpfile, err := ioutil.TempFile(dir, "manifest_parser_validate_runtimes_")
    if err == nil {
        defer os.Remove(tmpfile.Name()) // clean up
        if _, err := tmpfile.Write([]byte(data)); err == nil {
            // read and parse manifest.yaml file
            p := NewYAMLParser()
            m, _ := p.ParseManifest(tmpfile.Name())
            actions, err := p.ComposeActionsFromAllPackages(m, tmpfile.Name())
            var expectedResult string
            if err == nil {
                for i := 0; i < len(actions); i++ {
                    if actions[i].Action.Name == "helloNodejs" {
                        expectedResult = "nodejs:6"
                    } else if actions[i].Action.Name == "helloJava" {
                        expectedResult = "java"
                    } else if actions[i].Action.Name == "helloPython" {
                        expectedResult = "python"
                    } else if actions[i].Action.Name == "helloSwift" {
                        expectedResult = "swift:3"
                    }
                    actualResult := actions[i].Action.Exec.Kind
                    assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
                }
            }

        }
        tmpfile.Close()
    }
}

// Test 10: validate manifest_parser.ComposeActions() method for invalid runtimes
// when a runtime of an action is set to some garbage, manifest_parser should
// report an error for that action
func TestComposeActionsForInvalidRuntime(t *testing.T) {
    data :=
        `package:
   name: helloworld
   actions:
     helloInvalidRuntime:
       function: ../tests/src/integration/helloworld/actions/hello.js
       runtime: invalid`
    dir, _ := os.Getwd()
    tmpfile, err := ioutil.TempFile(dir, "manifest_parser_validate_runtime_")
    if err == nil {
        defer os.Remove(tmpfile.Name()) // clean up
        if _, err := tmpfile.Write([]byte(data)); err == nil {
            // read and parse manifest.yaml file
            p := NewYAMLParser()
            m, _ := p.ParseManifest(tmpfile.Name())
            _, err := p.ComposeActionsFromAllPackages(m, tmpfile.Name())
            // (TODO) uncomment the following test case after issue #307 is fixed
            // (TODO) its failing right now as we are lacking check on invalid runtime
            // TODO() https://github.com/apache/incubator-openwhisk-wskdeploy/issues/608
            // assert.NotNil(t, err, "Invalid runtime, ComposeActions should report an error")
            // (TODO) remove this print statement after uncommenting above test case
            fmt.Println(err)
        }
        tmpfile.Close()
    }
}

// Test 11: validate manfiest_parser.ComposeActions() method for single line parameters
// manifest_parser should be able to parse input section with different types of values
func TestComposeActionsForSingleLineParams(t *testing.T) {
    // manifest file is located under ../tests folder
    manifestFile := "../tests/dat/manifest_validate_singleline_params.yaml"
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(manifestFile)
    actions, err := p.ComposeActionsFromAllPackages(m, manifestFile)

    if err == nil {
        // assert that the actions variable has only one action
        assert.Equal(t, 1, len(actions), "We have defined only one action but we got " + string(len(actions)))

        action := actions[0]

        /*
         * Simple 'string' value tests
         */

        // param_simple_string should value "foo"
        expectedResult := "foo"
        actualResult := action.Action.Parameters.GetValue("param_simple_string").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        /*
         * Simple 'integer' value tests
         */

        // param_simple_integer_1 should have value 1
        expectedResult = strconv.FormatInt(1, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_simple_integer_1").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_integer_2 should have value 0
        expectedResult = strconv.FormatInt(0, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_simple_integer_2").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_integer_3 should have value -1
        expectedResult = strconv.FormatInt(-1, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_simple_integer_3").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_integer_4 should have value 99999
        expectedResult = strconv.FormatInt(99999, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_simple_integer_4").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_integer_5 should have value -99999
        expectedResult = strconv.FormatInt(-99999, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_simple_integer_5").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        /*
         * Simple 'float' value tests
         */

        // param_simple_float_1 should have value 1.1
        expectedResult = strconv.FormatFloat(1.1, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_simple_float_1").(float64), 'f', -1, 64)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_float_2 should have value 0.0
        expectedResult = strconv.FormatFloat(0.0, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_simple_float_2").(float64), 'f', -1, 64)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_float_3 should have value -1.1
        expectedResult = strconv.FormatFloat(-1.1, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_simple_float_3").(float64), 'f', -1, 64)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        /*
         * Environment Variable / dollar ($) notation tests
         */

        // param_simple_env_var_1 should have value of env. variable $GOPATH
        expectedResult = os.Getenv("GOPATH")
        actualResult = action.Action.Parameters.GetValue("param_simple_env_var_1").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_env_var_2 should have value of env. variable $GOPATH
        expectedResult = os.Getenv("GOPATH")
        actualResult = action.Action.Parameters.GetValue("param_simple_env_var_2").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_env_var_3 should have value of env. variable "${}"
        expectedResult = "${}"
        actualResult = action.Action.Parameters.GetValue("param_simple_env_var_3").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_invalid_env_var should have value of ""
        expectedResult = ""
        actualResult = action.Action.Parameters.GetValue("param_simple_invalid_env_var").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        /*
         * Environment Variable concatenation tests
         */

        // param_simple_env_var_concat_1 should have value of env. variable "$GOPTH/test" empty string
        expectedResult = os.Getenv("GOPATH") + "/test"
        actualResult = action.Action.Parameters.GetValue("param_simple_env_var_concat_1").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_env_var_concat_2 should have value of env. variable "" empty string
        // as the "/test" is treated as part of the environment var. and not concatenated.
        expectedResult = ""
        actualResult = action.Action.Parameters.GetValue("param_simple_env_var_concat_2").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_simple_env_var_concat_3 should have value of env. variable "" empty string
        expectedResult = "ddd.ccc." + os.Getenv("GOPATH")
        actualResult = action.Action.Parameters.GetValue("param_simple_env_var_concat_3").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        /*
         * Empty string tests
         */

        // param_simple_implied_empty should be ""
        actualResult = action.Action.Parameters.GetValue("param_simple_implied_empty").(string)
        assert.Empty(t, "", "Expected empty string but got " + actualResult)

        // param_simple_explicit_empty_1 should be ""
        actualResult = action.Action.Parameters.GetValue("param_simple_explicit_empty_1").(string)
        assert.Empty(t, "", "Expected empty string but got " + actualResult)

        // param_simple_explicit_empty_2 should be ""
        actualResult = action.Action.Parameters.GetValue("param_simple_explicit_empty_2").(string)
        assert.Empty(t, "", "Expected empty string but got " + actualResult)

        /*
         * Test values that contain "Type names" (e.g., "string", "integer", "float, etc.)
         */

        // param_simple_type_string should be "" when value set to "string"
        actualResult = action.Action.Parameters.GetValue("param_simple_type_string").(string)
        assert.Empty(t, "", "Expected empty string but got " + actualResult)

        // param_simple_type_integer should be 0.0 when value set to "integer"
        expectedResult = strconv.FormatInt(0, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_simple_type_integer").(int)), 10)
        assert.Empty(t, 0, "Expected empty string but got " + actualResult)

        // param_simple_type_float should be 0 when value set to "float"
        expectedResult = strconv.FormatFloat(0.0, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_simple_type_float").(float64), 'f', -1, 64)
        assert.Empty(t, 0.0, "Expected empty string but got " + actualResult)

    }
}

// Test 12: validate manfiest_parser.ComposeActions() method for multi line parameters
// manifest_parser should be able to parse input section with different types of values
func TestComposeActionsForMultiLineParams(t *testing.T) {
    // manifest file is located under ../tests folder
    manifestFile := "../tests/dat/manifest_validate_multiline_params.yaml"
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(manifestFile)
    actions, err := p.ComposeActionsFromAllPackages(m, manifestFile)

    if err == nil {
        // assert that the actions variable has only one action
        assert.Equal(t, 1, len(actions), "We have defined only one action but we got " + string(len(actions)))

        action := actions[0]

        fmt.Println(action.Action.Parameters)

        // param_string_value_only should be "foo"
        expectedResult := "foo"
        actualResult := action.Action.Parameters.GetValue("param_string_value_only").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_int_value_only should be 123
        expectedResult = strconv.FormatInt(123, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_int_value_only").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_float_value_only should be 3.14
        expectedResult = strconv.FormatFloat(3.14, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_float_value_only").(float64), 'f', -1, 64)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_string_type_and_value_only should be foo
        expectedResult = "foo"
        actualResult = action.Action.Parameters.GetValue("param_string_type_and_value_only").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_string_type_only should be ""
        actualResult = action.Action.Parameters.GetValue("param_string_type_only").(string)
        assert.Empty(t, actualResult, "Expected empty string but got " + actualResult)

        // param_integer_type_only should be 0
        expectedResult = strconv.FormatInt(0, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_integer_type_only").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_float_type_only should be 0
        expectedResult = strconv.FormatFloat(0.0, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_float_type_only").(float64), 'f', -1, 64)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_string_with_default should be "bar"
        expectedResult = "bar"
        actualResult = action.Action.Parameters.GetValue("param_string_with_default").(string)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_integer_with_default should be -1
        expectedResult = strconv.FormatInt(-1, 10)
        actualResult = strconv.FormatInt(int64(action.Action.Parameters.GetValue("param_integer_with_default").(int)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // param_float_with_default should be 2.9
        expectedResult = strconv.FormatFloat(2.9, 'f', -1, 64)
        actualResult = strconv.FormatFloat(action.Action.Parameters.GetValue("param_float_with_default").(float64), 'f', -1, 64)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
    }
}

// Test 13: validate manfiest_parser.ComposeActions() method
func TestComposeActionsForFunction(t *testing.T) {
    data :=
        `package:
  name: helloworld
  actions:
    hello1:
      function: ../tests/src/integration/helloworld/actions/hello.js`
    // (TODO) uncomment this after we add support for action file content from URL
    // hello2:
    //  function: https://raw.githubusercontent.com/apache/incubator-openwhisk-wskdeploy/master/tests/isrc/integration/helloworld/manifest.yaml`
    dir, _ := os.Getwd()
    tmpfile, err := ioutil.TempFile(dir, "manifest_parser_validate_locations_")
    if err == nil {
        defer os.Remove(tmpfile.Name()) // clean up
        if _, err := tmpfile.Write([]byte(data)); err == nil {
            // read and parse manifest.yaml file
            p := NewYAMLParser()
            m, _ := p.ParseManifest(tmpfile.Name())
            actions, err := p.ComposeActionsFromAllPackages(m, tmpfile.Name())
            var expectedResult, actualResult string
            if err == nil {
                for i := 0; i < len(actions); i++ {
                    if actions[i].Action.Name == "hello1" {
                        expectedResult, _ = filepath.Abs("../tests/src/integration/helloworld/actions/hello.js")
                        actualResult, _ = filepath.Abs(actions[i].Filepath)
                        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)
                        // (TODO) Uncomment the following condition, hello2
                        // (TODO) after issue # 311 is fixed
                        //} else if actions[i].Action.Name == "hello2" {
                        //  assert.NotNil(t, actions[i].Action.Exec.Code, "Expected source code from an action file but found it empty")
                    }
                }
            }

        }
        tmpfile.Close()
    }

}

// Test 14: validate manfiest_parser.ComposeActions() method
func TestComposeActionsForLimits (t *testing.T) {
  data :=
`package:
  name: helloworld
  actions:
    hello1:
      function: ../tests/src/integration/helloworld/actions/hello.js
      limits:
        timeout: 1
    hello2:
      function: ../tests/src/integration/helloworld/actions/hello.js
      limits:
        timeout: 180
        memorySize: 128
        logSize: 1
        concurrentActivations: 10
        userInvocationRate: 50
        codeSize: 1024
        parameterSize: 128`
  dir, _ := os.Getwd()
  tmpfile, err := ioutil.TempFile(dir, "manifest_parser_validate_limits_")
  if err == nil {
      defer os.Remove(tmpfile.Name()) // clean up
      if _, err := tmpfile.Write([]byte(data)); err == nil {
          // read and parse manifest.yaml file
          p := NewYAMLParser()
          m, _ := p.ParseManifest(tmpfile.Name())
          actions, err := p.ComposeActionsFromAllPackages(m, tmpfile.Name())
          //var expectedResult, actualResult string
          if err == nil {
              for i:=0; i<len(actions); i++ {
                  if actions[i].Action.Name == "hello1" {
                      assert.Nil(t, actions[i].Action.Limits, "Expected limit section to be empty but got %s", actions[i].Action.Limits)
                  } else if actions[i].Action.Name == "hello2" {
                      assert.NotNil(t, actions[i].Action.Limits, "Expected limit section to be not empty but found it empty")
                      assert.Equal(t, 180, *actions[i].Action.Limits.Timeout, "Failed to get Timeout")
                      assert.Equal(t, 128, *actions[i].Action.Limits.Memory, "Failed to get Memory")
                      assert.Equal(t, 1, *actions[i].Action.Limits.Logsize, "Failed to get Logsize")
                  }
              }
          }

      }
      tmpfile.Close()
  }
}

// Test 15: validate manfiest_parser.ComposeActions() method
func TestComposeActionsForWebActions(t *testing.T) {
    data :=
        `package:
  name: helloworld
  actions:
    hello:
      function: ../tests/src/integration/helloworld/actions/hello.js
      web-export: true`
    dir, _ := os.Getwd()
    tmpfile, err := ioutil.TempFile(dir, "manifest_parser_validate_web_actions_")
    if err == nil {
        defer os.Remove(tmpfile.Name()) // clean up
        if _, err := tmpfile.Write([]byte(data)); err == nil {
            // read and parse manifest.yaml file
            p := NewYAMLParser()
            m, _ := p.ParseManifest(tmpfile.Name())
            actions, err := p.ComposeActionsFromAllPackages(m, tmpfile.Name())
            if err == nil {
                for i := 0; i < len(actions); i++ {
                    if actions[i].Action.Name == "hello" {
                        for _, a := range actions[i].Action.Annotations {
                            switch a.Key {
                            case "web-export":
                                assert.Equal(t, true, a.Value, "Expected true for web-export but got " + strconv.FormatBool(a.Value.(bool)))
                            case "raw-http":
                                assert.Equal(t, false, a.Value, "Expected false for raw-http but got " + strconv.FormatBool(a.Value.(bool)))
                            case "final":
                                assert.Equal(t, true, a.Value, "Expected true for final but got " + strconv.FormatBool(a.Value.(bool)))
                            }
                        }
                    }
                }
            }

        }
        tmpfile.Close()
    }
}

// Test 16: validate manifest_parser.ResolveParameter() method
func TestResolveParameterForMultiLineParams(t *testing.T) {
    p := "name"
    v := "foo"
    y := reflect.TypeOf(v).Name() // y := string
    d := "default_name"

    // type string - value only param
    param1 := Parameter{Value: v, multiline: true}
    r1, _ := ResolveParameter(p, &param1, "")
    assert.Equal(t, v, r1, "Expected value " + v + " but got " + r1.(string))
    assert.IsType(t, v, r1, "Expected parameter %v of type %T but found %T", p, v, r1)

    // type string - type and value only param
    param2 := Parameter{Type: y, Value: v, multiline: true}
    r2, _ := ResolveParameter(p, &param2, "")
    assert.Equal(t, v, r2, "Expected value " + v + " but got " + r2.(string))
    assert.IsType(t, v, r2, "Expected parameter %v of type %T but found %T", p, v, r2)

    // type string - type, no value, but default value param
    param3 := Parameter{Type: y, Default: d, multiline: true}
    r3, _ := ResolveParameter(p, &param3, "")
    assert.Equal(t, d, r3, "Expected value " + d + " but got " + r3.(string))
    assert.IsType(t, d, r3, "Expected parameter %v of type %T but found %T", p, d, r3)

    // type string - type and value only param
    // type is "string" and value is of type "int"
    // ResolveParameter matches specified type with the type of the specified value
    // it fails if both types don't match
    // ResolveParameter determines type from the specified value
    // in this case, ResolveParameter returns value of type int
    v1 := 11
    param4 := Parameter{Type: y, Value: v1, multiline: true}
    r4, _ := ResolveParameter(p, &param4, "")
    assert.Equal(t, v1, r4, "Expected value " + strconv.FormatInt(int64(v1), 10) + " but got " + strconv.FormatInt(int64(r4.(int)), 10))
    assert.IsType(t, v1, r4, "Expected parameter %v of type %T but found %T", p, v1, r4)

    // type invalid - type only param
    param5 := Parameter{Type: "invalid", multiline: true}
    _, err := ResolveParameter(p, &param5, "")
    assert.NotNil(t, err, "Expected error saying Invalid type for parameter")
    lines := []string{"Line Unknown"}
    msgs := []string{"Invalid Type for parameter. [invalid]"}
    expectedErr := utils.NewParserErr("", lines, msgs)
    switch errorType := err.(type) {
    default:
        assert.Fail(t, "Wrong error type received: We are expecting ParserErr.")
    case *utils.ParserErr:
        assert.Equal(t, expectedErr.Message, errorType.Message,
            "Expected error " + expectedErr.Message + " but found " + errorType.Message)
    }

    // type none - param without type, without value, and without default value
    param6 := Parameter{multiline: true}
    r6, _ := ResolveParameter("none", &param6, "")
    assert.Empty(t, r6, "Expected default value of empty string but found " + r6.(string))

}

// Test 17: validate JSON parameters
func TestParseManifestForJSONParams(t *testing.T) {
    // manifest file is located under ../tests folder
    manifestFile := "../tests/dat/manifest_validate_json_params.yaml"
    // read and parse manifest.yaml file
    m, _ := NewYAMLParser().ParseManifest(manifestFile)

    // validate package name should be "validate"
    packageName := "validate_json"
    actionName := "validate_json_params"
    expectedActionsCount := 1

    assert.NotNil(t, m.Packages[packageName],
        "Expected package named "+ packageName + " but got none")

    // validate this package contains one action
    actualActionsCount := len(m.Packages[packageName].Actions)
    assert.Equal(t, expectedActionsCount, actualActionsCount,
        "Expected " + string(expectedActionsCount) + " but got " + string(actualActionsCount))

    if action, ok := m.Packages[packageName].Actions[actionName]; ok {
        // validate location/function of an action to be "actions/dump_params.js"
        expectedResult := "actions/dump_params.js"
        actualResult := action.Function
        assert.Equal(t, expectedResult, actualResult, "Expected action function " + expectedResult + " but got " + actualResult)

        // validate runtime of an action to be "nodejs:6"
        expectedResult = "nodejs:6"
        actualResult = action.Runtime
        assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)

        // validate the number of inputs to this action
        expectedResult = strconv.FormatInt(6, 10)
        actualResult = strconv.FormatInt(int64(len(action.Inputs)), 10)
        assert.Equal(t, expectedResult, actualResult, "Expected " + expectedResult + " but got " + actualResult)

        // validate inputs to this action
        for input, param := range action.Inputs {
            // Trace to help debug complex values:
            // utils.PrintTypeInfo(input, param.Value)
            switch input {
            case "member1":
                actualResult1 := param.Value.(string)
                expectedResult1 := "{ \"name\": \"Sam\", \"place\": \"Shire\" }"
                assert.Equal(t, expectedResult1, actualResult1, "Expected " + expectedResult + " but got " + actualResult)
            case "member2":
                actualResult2 := param.Value.(map[interface{}]interface{})
                expectedResult2 := map[interface{}]interface{}{"name": "Sam", "place": "Shire"}
                assert.Equal(t, expectedResult2, actualResult2, "Expected " + expectedResult + " but got " + actualResult)
            case "member3":
                actualResult3 := param.Value.(map[interface{}]interface{})
                expectedResult3 := map[interface{}]interface{}{"name": "Elrond", "place": "Rivendell"}
                assert.Equal(t, expectedResult3, actualResult3, "Expected " + expectedResult + " but got " + actualResult)
            case "member4":
                actualResult4 := param.Value.(map[interface{}]interface{})
                expectedResult4 := map[interface{}]interface{}{"name": "Gimli", "place": "Gondor", "age": 139, "children": map[interface{}]interface{}{ "<none>": "<none>" }}
                assert.Equal(t, expectedResult4, actualResult4, "Expected " + expectedResult + " but got " + actualResult)
            case "member5":
                actualResult5 := param.Value.(map[interface{}]interface{})
                expectedResult5 := map[interface{}]interface{}{"name": "Gloin", "place": "Gondor", "age": 235, "children": map[interface{}]interface{}{ "Gimli": "Son" }}
                assert.Equal(t, expectedResult5, actualResult5, "Expected " + expectedResult + " but got " + actualResult)
            case "member6":
                actualResult6 := param.Value.(map[interface{}]interface{})
                expectedResult6 := map[interface{}]interface{}{"name": "Frodo", "place": "Undying Lands", "items": []interface{}{"Sting", "Mithril mail"}}
                assert.Equal(t, expectedResult6, actualResult6, "Expected " + expectedResult + " but got " + actualResult)
            }
        }

        // validate Outputs from this action
        for output, param := range action.Outputs {
            switch output {
            case "fellowship":
                expectedResultA := "json"
                actualResultA := param.Type
                assert.Equal(t, expectedResultA, actualResultA, "Expected " + expectedResultA + " but got " + actualResultA)

                expectedResultB := map[interface{}]interface{}{}
                actualResultB := param.Value
                //actualResultB := reflect.TypeOf(expectedResult).String()
                fmt.Printf("exp=%s, act=%s", reflect.TypeOf(expectedResult).String(),reflect.TypeOf(expectedResult).String() )
                //assert.Equal(t, expectedResultB, actualResultB, "Expected " + expectedResultB + " but got " + actualResultB)
            }
        }
    }
}

func _createTmpfile(data string, filename string) (f *os.File, err error) {
    dir, _ := os.Getwd()
    tmpfile, err := ioutil.TempFile(dir, filename)
    if err != nil {
        return nil, err
    }
    _, err = tmpfile.Write([]byte(data))
    if err != nil {
        return tmpfile, err
    }
    return tmpfile, nil
}

func TestComposePackage(t *testing.T) {
    data := `package:
  name: helloworld
  namespace: default`
    tmpfile, err := _createTmpfile(data, "manifest_parser_test_compose_package_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(tmpfile.Name())
    pkg, err := p.ComposeAllPackages(m, tmpfile.Name())
    if err == nil {
        n := "helloworld"
        assert.NotNil(t, pkg[n], "Failed to get the whole package")
        assert.Equal(t, n, pkg[n].Name, "Failed to get package name")
        assert.Equal(t, "default", pkg[n].Namespace, "Failed to get package namespace")
    } else {
        assert.Fail(t, "Failed to compose package")
    }
}

func TestComposeSequences(t *testing.T) {
    data := `package:
  name: helloworld
  sequences:
    sequence1:
      actions: action1, action2
    sequence2:
      actions: action3, action4, action5`
    tmpfile, err := _createTmpfile(data, "manifest_parser_test_compose_package_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(tmpfile.Name())
    seqList, err := p.ComposeSequencesFromAllPackages("", m)
    if err != nil {
        assert.Fail(t, "Failed to compose sequences")
    }
    assert.Equal(t, 2, len(seqList), "Failed to get sequences")
    for _, seq := range seqList {
        wsk_action := seq.Action
        switch wsk_action.Name {
        case "sequence1":
            assert.Equal(t, "sequence", wsk_action.Exec.Kind, "Failed to set sequence exec kind")
            assert.Equal(t, 2, len(wsk_action.Exec.Components), "Failed to set sequence exec components")
            assert.Equal(t, "/helloworld/action1", wsk_action.Exec.Components[0], "Failed to set sequence 1st exec components")
            assert.Equal(t, "/helloworld/action2", wsk_action.Exec.Components[1], "Failed to set sequence 2nd exec components")
        case "sequence2":
            assert.Equal(t, "sequence", wsk_action.Exec.Kind, "Failed to set sequence exec kind")
            assert.Equal(t, 3, len(wsk_action.Exec.Components), "Failed to set sequence exec components")
            assert.Equal(t, "/helloworld/action3", wsk_action.Exec.Components[0], "Failed to set sequence 1st exec components")
            assert.Equal(t, "/helloworld/action4", wsk_action.Exec.Components[1], "Failed to set sequence 2nd exec components")
            assert.Equal(t, "/helloworld/action5", wsk_action.Exec.Components[2], "Failed to set sequence 3rd exec components")
        }
    }
}

func TestComposeTriggers(t *testing.T) {
    data := `package:
  name: helloworld
  triggers:
    trigger1:
      inputs:
        name: string
        place: string
    trigger2:
      feed: myfeed
      inputs:
        name: myname
        place: myplace`
    tmpfile, err := _createTmpfile(data, "manifest_parser_test_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(tmpfile.Name())
    triggerList, err := p.ComposeTriggersFromAllPackages(m, tmpfile.Name())
    if err != nil {
        assert.Fail(t, "Failed to compose trigger")
    }

    assert.Equal(t, 2, len(triggerList), "Failed to get trigger list")
    for _, trigger := range triggerList {
        switch trigger.Name {
        case "trigger1":
            assert.Equal(t, 2, len(trigger.Parameters), "Failed to set trigger parameters")
        case "trigger2":
            assert.Equal(t, "feed", trigger.Annotations[0].Key, "Failed to set trigger annotation")
            assert.Equal(t, "myfeed", trigger.Annotations[0].Value, "Failed to set trigger annotation")
            assert.Equal(t, 2, len(trigger.Parameters), "Failed to set trigger parameters")
        }
    }
}

func TestComposeRules(t *testing.T) {
    data := `package:
  name: helloworld
  rules:
    rule1:
      trigger: locationUpdate
      action: greeting
    rule2:
      trigger: trigger1
      action: action1`
    tmpfile, err := _createTmpfile(data, "manifest_parser_test_compose_package_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(tmpfile.Name())
    ruleList, err := p.ComposeRulesFromAllPackages(m)
    if err != nil {
        assert.Fail(t, "Failed to compose rules")
    }
    assert.Equal(t, 2, len(ruleList), "Failed to get rules")
    for _, rule := range ruleList {
        switch rule.Name {
        case "rule1":
            assert.Equal(t, "locationUpdate", rule.Trigger, "Failed to set rule trigger")
            assert.Equal(t, "helloworld/greeting", rule.Action, "Failed to set rule action")
        case "rule2":
            assert.Equal(t, "trigger1", rule.Trigger, "Failed to set rule trigger")
            assert.Equal(t, "helloworld/action1", rule.Action, "Failed to set rule action")
        }
    }
}

func TestComposeApiRecords(t *testing.T) {
    data := `package:
  name: helloworld
  apis:
    book-club:
      club:
        books:
           putBooks: put
           deleteBooks: delete
        members:
           listMembers: get
    book-club2:
      club2:
        books2:
           getBooks2: get
           postBooks2: post
        members2:
           listMembers2: get`
    tmpfile, err := _createTmpfile(data, "manifest_parser_test_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(tmpfile.Name())
    apiList, err := p.ComposeApiRecordsFromAllPackages(m)
    if err != nil {
        assert.Fail(t, "Failed to compose api records")
    }
    assert.Equal(t, 6, len(apiList), "Failed to get api records")
    for _, apiRecord := range apiList {
        apiDoc := apiRecord.ApiDoc
        action := apiDoc.Action
        switch action.Name {
        case "putBooks":
            assert.Equal(t, "book-club", apiDoc.ApiName, "Failed to set api name")
            assert.Equal(t, "club", apiDoc.GatewayBasePath, "Failed to set api base path")
            assert.Equal(t, "books", apiDoc.GatewayRelPath, "Failed to set api rel path")
            assert.Equal(t, "put", action.BackendMethod, "Failed to set api backend method")
        case "deleteBooks":
            assert.Equal(t, "book-club", apiDoc.ApiName, "Failed to set api name")
            assert.Equal(t, "club", apiDoc.GatewayBasePath, "Failed to set api base path")
            assert.Equal(t, "books", apiDoc.GatewayRelPath, "Failed to set api rel path")
            assert.Equal(t, "delete", action.BackendMethod, "Failed to set api backend method")
        case "listMembers":
            assert.Equal(t, "book-club", apiDoc.ApiName, "Failed to set api name")
            assert.Equal(t, "club", apiDoc.GatewayBasePath, "Failed to set api base path")
            assert.Equal(t, "members", apiDoc.GatewayRelPath, "Failed to set api rel path")
            assert.Equal(t, "get", action.BackendMethod, "Failed to set api backend method")
        case "getBooks2":
            assert.Equal(t, "book-club2", apiDoc.ApiName, "Failed to set api name")
            assert.Equal(t, "club2", apiDoc.GatewayBasePath, "Failed to set api base path")
            assert.Equal(t, "books2", apiDoc.GatewayRelPath, "Failed to set api rel path")
            assert.Equal(t, "get", action.BackendMethod, "Failed to set api backend method")
        case "postBooks2":
            assert.Equal(t, "book-club2", apiDoc.ApiName, "Failed to set api name")
            assert.Equal(t, "club2", apiDoc.GatewayBasePath, "Failed to set api base path")
            assert.Equal(t, "books2", apiDoc.GatewayRelPath, "Failed to set api rel path")
            assert.Equal(t, "post", action.BackendMethod, "Failed to set api backend method")
        case "listMembers2":
            assert.Equal(t, "book-club2", apiDoc.ApiName, "Failed to set api name")
            assert.Equal(t, "club2", apiDoc.GatewayBasePath, "Failed to set api base path")
            assert.Equal(t, "members2", apiDoc.GatewayRelPath, "Failed to set api rel path")
            assert.Equal(t, "get", action.BackendMethod, "Failed to set api backend method")
        default:
            assert.Fail(t, "Failed to get api action name")
        }
    }
}

func TestComposeDependencies(t *testing.T) {
    data := `package:
  name: helloworld
  dependencies:
    myhelloworld:
      location: github.com/user/repo/folder
    myCloudant:
      location: /whisk.system/cloudant
      inputs:
        dbname: myGreatDB
      annotations:
        myAnnotation: Here it is`
    tmpfile, err := _createTmpfile(data, "manifest_parser_test_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    // read and parse manifest.yaml file
    p := NewYAMLParser()
    m, _ := p.ParseManifest(tmpfile.Name())
    depdList, err := p.ComposeDependenciesFromAllPackages(m, "/project_folder", tmpfile.Name())
    if err != nil {
        assert.Fail(t, "Failed to compose rules")
    }
    assert.Equal(t, 2, len(depdList), "Failed to get rules")
    for depdy_name, depdy := range depdList {
        assert.Equal(t, "helloworld", depdy.Packagename, "Failed to set dependecy isbinding")
        assert.Equal(t, "/project_folder/Packages", depdy.ProjectPath, "Failed to set dependecy isbinding")
        d := strings.Split(depdy_name, ":")
        assert.NotEqual(t, d[1], "", "Failed to get dependency name")
        switch d[1] {
        case "myhelloworld":
            assert.Equal(t, "https://github.com/user/repo/folder", depdy.Location, "Failed to set dependecy location")
            assert.Equal(t, false, depdy.IsBinding, "Failed to set dependecy isbinding")
            assert.Equal(t, "https://github.com/user/repo", depdy.BaseRepo, "Failed to set dependecy base repo url")
            assert.Equal(t, "/folder", depdy.SubFolder, "Failed to set dependecy sub folder")
        case "myCloudant":
            assert.Equal(t, "/whisk.system/cloudant", depdy.Location, "Failed to set rule trigger")
            assert.Equal(t, true, depdy.IsBinding, "Failed to set dependecy isbinding")
            assert.Equal(t, 1, len(depdy.Parameters), "Failed to set dependecy parameter")
            assert.Equal(t, 1, len(depdy.Annotations), "Failed to set dependecy annotation")
            assert.Equal(t, "myAnnotation", depdy.Annotations[0].Key, "Failed to set dependecy parameter key")
            assert.Equal(t, "Here it is", depdy.Annotations[0].Value, "Failed to set dependecy parameter value")
            assert.Equal(t, "dbname", depdy.Parameters[0].Key, "Failed to set dependecy annotation key")
            assert.Equal(t, "myGreatDB", depdy.Parameters[0].Value, "Failed to set dependecy annotation value")
        default:
            assert.Fail(t, "Failed to get dependency name")
        }
    }
}

func TestBadYAMLInvalidPackageKeyInManifest(t *testing.T) {
    // read and parse manifest.yaml file located under ../tests folder
    p := NewYAMLParser()
    _, err := p.ParseManifest("../tests/dat/manifest_bad_yaml_invalid_package_key.yaml")

    assert.NotNil(t, err)
    // go-yaml/yaml prints the wrong line number for mapping values. It should be 4.
    assert.Contains(t, err.Error(), "line 2: field invalidKey not found in struct parsers.Package")
}

func TestBadYAMLInvalidKeyMappingValueInManifest(t *testing.T) {
    // read and parse manifest.yaml file located under ../tests folder
    p := NewYAMLParser()
    _, err := p.ParseManifest("../tests/dat/manifest_bad_yaml_invalid_key_mapping_value.yaml")

    assert.NotNil(t, err)
    // go-yaml/yaml prints the wrong line number for mapping values. It should be 5.
    assert.Contains(t, err.Error(), "line 4: mapping values are not allowed in this context")
}

func TestBadYAMLMissingRootKeyInManifest(t *testing.T) {
    // read and parse manifest.yaml file located under ../tests folder
    p := NewYAMLParser()
    _, err := p.ParseManifest("../tests/dat/manifest_bad_yaml_missing_root_key.yaml")

    assert.NotNil(t, err)
    assert.Contains(t, err.Error(), "line 1: field actions not found in struct parsers.YAML")
}

func TestBadYAMLInvalidCommentInManifest(t *testing.T) {
    // read and parse manifest.yaml file located under ../tests folder
    p := NewYAMLParser()
    _, err := p.ParseManifest("../tests/dat/manifest_bad_yaml_invalid_comment.yaml")

    assert.NotNil(t, err)
    assert.Contains(t, err.Error(), "line 13: could not find expected ':'")
}

// validate manifest_parser:Unmarshal() method for package in manifest YAML
// validate that manifest_parser is able to read and parse the manifest data
func TestUnmarshalForPackages(t *testing.T) {
    data := `
packages:
  package1:
    actions:
      helloNodejs:
        function: actions/hello.js
        runtime: nodejs:6
  package2:
    actions:
      helloPython:
        function: actions/hello.py
        runtime: python`
    // set the zero value of struct YAML
    m := YAML{}
    // Unmarshal reads/parses manifest data and sets the values of YAML
    // And returns an error if parsing a manifest data fails
    err := NewYAMLParser().Unmarshal([]byte(data), &m)
    if err == nil {
        expectedResult := string(2)
        actualResult := string(len(m.Packages))
        assert.Equal(t, expectedResult, actualResult, "Expected 2 packages but got " + actualResult)
        // we have two packages
        // package name should be "helloNodejs" and "helloPython"
        for k, v := range m.Packages {
            switch k {
            case "package1":
                assert.Equal(t, "package1", k, "Expected package name package1 but got " + k)
                expectedResult = string(1)
                actualResult = string(len(v.Actions))
                assert.Equal(t, expectedResult, actualResult, "Expected 1 but got " + actualResult)
                // get the action payload from the map of actions which is stored in
                // YAML.Package.Actions with the type of map[string]Action
                actionName := "helloNodejs"
                if action, ok := v.Actions[actionName]; ok {
                    // location/function of an action should be "actions/hello.js"
                    expectedResult = "actions/hello.js"
                    actualResult = action.Function
                    assert.Equal(t, expectedResult, actualResult, "Expected action function " + expectedResult + " but got " + actualResult)
                    // runtime of an action should be "nodejs:6"
                    expectedResult = "nodejs:6"
                    actualResult = action.Runtime
                    assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)
                } else {
                    t.Error("Action named " + actionName + " does not exist.")
                }
            case "package2":
                assert.Equal(t, "package2", k, "Expected package name package2 but got " + k)
                expectedResult = string(1)
                actualResult = string(len(v.Actions))
                assert.Equal(t, expectedResult, actualResult, "Expected 1 but got " + actualResult)
                // get the action payload from the map of actions which is stored in
                // YAML.Package.Actions with the type of map[string]Action
                actionName := "helloPython"
                if action, ok := v.Actions[actionName]; ok {
                    // location/function of an action should be "actions/hello.js"
                    expectedResult = "actions/hello.py"
                    actualResult = action.Function
                    assert.Equal(t, expectedResult, actualResult, "Expected action function " + expectedResult + " but got " + actualResult)
                    // runtime of an action should be "python"
                    expectedResult = "python"
                    actualResult = action.Runtime
                    assert.Equal(t, expectedResult, actualResult, "Expected action runtime " + expectedResult + " but got " + actualResult)
                } else {
                    t.Error("Action named " + actionName + " does not exist.")
                }
            }
        }
    }
}

func TestParseYAML_trigger(t *testing.T) {
	data, err := ioutil.ReadFile("../tests/dat/manifest_validate_triggerfeed.yaml")
	if err != nil {
		panic(err)
	}

	var manifest YAML
	err = NewYAMLParser().Unmarshal(data, &manifest)
	if err != nil {
		panic(err)
	}

        packageName := "manifest3"

	assert.Equal(t, 2, len(manifest.Packages[packageName].Triggers), "Get trigger list failed.")
	for trigger_name := range manifest.Packages[packageName].Triggers {
		var trigger = manifest.Packages[packageName].Triggers[trigger_name]
		switch trigger_name {
		case "trigger1":
		case "trigger2":
			assert.Equal(t, "myfeed", trigger.Feed, "Get trigger feed name failed.")
		default:
			t.Error("Get trigger name failed")
		}
	}
}

func TestParseYAML_rule(t *testing.T) {
	data, err := ioutil.ReadFile("../tests/dat/manifest_validate_rule.yaml")
	if err != nil {
		panic(err)
	}

	var manifest YAML
	err = NewYAMLParser().Unmarshal(data, &manifest)
	if err != nil {
		panic(err)
	}

        packageName := "manifest4"

	assert.Equal(t, 1, len(manifest.Packages[packageName].Rules), "Get trigger list failed.")
	for rule_name := range manifest.Packages[packageName].Rules {
		var rule = manifest.Packages[packageName].Rules[rule_name]
		switch rule_name {
		case "rule1":
			assert.Equal(t, "trigger1", rule.Trigger, "Get trigger name failed.")
			assert.Equal(t, "hellpworld", rule.Action, "Get action name failed.")
			assert.Equal(t, "true", rule.Rule, "Get rule expression failed.")
		default:
			t.Error("Get rule name failed")
		}
	}
}

func TestParseYAML_feed(t *testing.T) {
	data, err := ioutil.ReadFile("../tests/dat/manifest_validate_feed.yaml")
	if err != nil {
		panic(err)
	}

	var manifest YAML
	err = NewYAMLParser().Unmarshal(data, &manifest)
	if err != nil {
		panic(err)
	}

        packageName := "manifest5"

	assert.Equal(t, 1, len(manifest.Packages[packageName].Feeds), "Get feed list failed.")
	for feed_name := range manifest.Packages[packageName].Feeds {
		var feed = manifest.Packages[packageName].Feeds[feed_name]
		switch feed_name {
		case "feed1":
			assert.Equal(t, "https://my.company.com/services/eventHub", feed.Location, "Get feed location failed.")
			assert.Equal(t, "my_credential", feed.Credential, "Get feed credential failed.")
			assert.Equal(t, 2, len(feed.Operations), "Get operations number failed.")
			for operation_name := range feed.Operations {
				switch operation_name {
				case "operation1":
				case "operation2":
				default:
					t.Error("Get feed operation name failed")
				}
			}
		default:
			t.Error("Get feed name failed")
		}
	}
}

func TestParseYAML_param(t *testing.T) {
	data, err := ioutil.ReadFile("../tests/dat/manifest_validate_params.yaml")
	if err != nil {
		panic(err)
	}

	var manifest YAML
	err = NewYAMLParser().Unmarshal(data, &manifest)
	if err != nil {
		panic(err)
	}

        packageName := "validateParams"

	assert.Equal(t, 1, len(manifest.Packages[packageName].Actions), "Get action list failed.")
	for action_name := range manifest.Packages[packageName].Actions {
		var action = manifest.Packages[packageName].Actions[action_name]
		switch action_name {
		case "action1":
			for param_name := range action.Inputs {
				var param = action.Inputs[param_name]
				switch param_name {
				case "inline1":
					assert.Equal(t, "{ \"key\": true }", param.Value, "Get param value failed.")
				case "inline2":
					assert.Equal(t, "Just a string", param.Value, "Get param value failed.")
				case "inline3":
					assert.Equal(t, nil, param.Value, "Get param value failed.")
				case "inline4":
					assert.Equal(t, true, param.Value, "Get param value failed.")
				case "inline5":
					assert.Equal(t, 42, param.Value, "Get param value failed.")
				case "inline6":
					assert.Equal(t, -531, param.Value, "Get param value failed.")
				case "inline7":
					assert.Equal(t, 432.432E-43, param.Value, "Get param value failed.")
				case "inline8":
					assert.Equal(t, "[ true, null, \"boo\", { \"key\": 0 }]", param.Value, "Get param value failed.")
				case "inline9":
					assert.Equal(t, false, param.Value, "Get param value failed.")
				case "inline0":
					assert.Equal(t, 456.423, param.Value, "Get param value failed.")
				case "inlin10":
					assert.Equal(t, nil, param.Value, "Get param value failed.")
				case "inlin11":
					assert.Equal(t, true, param.Value, "Get param value failed.")
				case "expand1":
					assert.Equal(t, nil, param.Value, "Get param value failed.")
				case "expand2":
					assert.Equal(t, true, param.Value, "Get param value failed.")
				case "expand3":
					assert.Equal(t, false, param.Value, "Get param value failed.")
				case "expand4":
					assert.Equal(t, 15646, param.Value, "Get param value failed.")
				case "expand5":
					assert.Equal(t, "{ \"key\": true }", param.Value, "Get param value failed.")
				case "expand6":
					assert.Equal(t, "[ true, null, \"boo\", { \"key\": 0 }]", param.Value, "Get param value failed.")
				case "expand7":
					assert.Equal(t, nil, param.Value, "Get param value failed.")
				default:
					t.Error("Get param name failed")
				}
			}
		default:
			t.Error("Get action name failed")
		}
	}
}
