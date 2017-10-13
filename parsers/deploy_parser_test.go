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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
    "io/ioutil"
)

func createTmpfile(data string, filename string) (f *os.File, err error) {
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

func TestInvalidKeyDeploymentYaml(t *testing.T) {
    data :=`application:
  name: wskdeploy-samples
  invalidKey: test`
    tmpfile, err := createTmpfile(data, "deployment_parser_test_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    p := NewYAMLParser()
    _, err = p.ParseDeployment(tmpfile.Name())
    assert.NotNil(t, err)
    // go-yaml/yaml prints the wrong line number for mapping values. It should be 3.
    assert.Contains(t, err.Error(), "line 2: field invalidKey not found in struct parsers.Application")
}

func TestMappingValueDeploymentYaml(t *testing.T) {
    data :=`application:
  name: wskdeploy-samples
    packages: test`
    tmpfile, err := createTmpfile(data, "deployment_parser_test_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    p := NewYAMLParser()
    _, err = p.ParseDeployment(tmpfile.Name())
    assert.NotNil(t, err)
    // go-yaml/yaml prints the wrong line number for mapping values. It should be 3.
    assert.Contains(t, err.Error(), "yaml: line 2: mapping values are not allowed in this context")
}

func TestMissingRootNodeDeploymentYaml(t *testing.T) {
    data :=`name: wskdeploy-samples`
    tmpfile, err := createTmpfile(data, "deployment_parser_test_")
    if err != nil {
        assert.Fail(t, "Failed to create temp file")
    }
    defer func() {
        tmpfile.Close()
        os.Remove(tmpfile.Name())
    }()
    p := NewYAMLParser()
    _, err = p.ParseDeployment(tmpfile.Name())
    assert.NotNil(t, err)
    // go-yaml/yaml prints the wrong line number for mapping values. It should be 3.
    assert.Contains(t, err.Error(), "line 1: field name not found in struct parsers.YAML")
}

func TestParseDeploymentYAML_Application(t *testing.T) {
	//var deployment utils.DeploymentYAML
	mm := NewYAMLParser()
	deployment, _ := mm.ParseDeployment("../tests/dat/deployment_data_application.yaml")

	//get and verify application name
	assert.Equal(t, "wskdeploy-samples", deployment.Application.Name, "Get application name failed.")
	assert.Equal(t, "/wskdeploy/samples/", deployment.Application.Namespace, "Get application namespace failed.")
	assert.Equal(t, "user-credential", deployment.Application.Credential, "Get application credential failed.")
	assert.Equal(t, "172.17.0.1", deployment.Application.ApiHost, "Get application api host failed.")
}

func TestParseDeploymentYAML_Application_Package(t *testing.T) {
	//var deployment utils.DeploymentYAML
	mm := NewYAMLParser()
	deployment, _ := mm.ParseDeployment("../tests/dat/deployment_data_application_package.yaml")

	assert.Equal(t, 1, len(deployment.Application.Packages), "Get package list failed.")
	for pkg_name := range deployment.Application.Packages {
		assert.Equal(t, "test_package", pkg_name, "Get package name failed.")
		var pkg = deployment.Application.Packages[pkg_name]
		assert.Equal(t, "/wskdeploy/samples/test", pkg.Namespace, "Get package namespace failed.")
		assert.Equal(t, "12345678ABCDEF", pkg.Credential, "Get package credential failed.")
		assert.Equal(t, 1, len(pkg.Inputs), "Get package input list failed.")
		//get and verify inputs
		for param_name, param := range pkg.Inputs {
			assert.Equal(t, "value", param.Value, "Get input value failed.")
			assert.Equal(t, "param", param_name, "Get input param name failed.")
		}
	}
}

func TestParseDeploymentYAML_Packages(t *testing.T) {
    //var deployment utils.DeploymentYAML
    mm := NewYAMLParser()
    deployment, _ := mm.ParseDeployment("../tests/dat/deployment_data_packages.yaml")

    assert.Equal(t, 0, len(deployment.Application.Packages), "Packages under application are empty.")
    assert.Equal(t, 0, len(deployment.Application.Package.Packagename), "Package name is empty.")
    assert.Equal(t, 1, len(deployment.Packages), "Packages are available.")
    for pkg_name := range deployment.Packages {
        assert.Equal(t, "test_package", pkg_name, "Get package name failed.")
        var pkg = deployment.Packages[pkg_name]
        assert.Equal(t, "/wskdeploy/samples/test", pkg.Namespace, "Get package namespace failed.")
        assert.Equal(t, "12345678ABCDEF", pkg.Credential, "Get package credential failed.")
        assert.Equal(t, 1, len(pkg.Inputs), "Get package input list failed.")
        //get and verify inputs
        for param_name, param := range pkg.Inputs {
            assert.Equal(t, "value", param.Value, "Get input value failed.")
            assert.Equal(t, "param", param_name, "Get input param name failed.")
        }
    }
}

func TestParseDeploymentYAML_Package(t *testing.T) {
    //var deployment utils.DeploymentYAML
    mm := NewYAMLParser()
    deployment, _ := mm.ParseDeployment("../tests/dat/deployment_data_package.yaml")

    assert.Equal(t, 0, len(deployment.Application.Packages), "Get package list failed.")
    assert.Equal(t, 0, len(deployment.Application.Package.Packagename), "Package name is empty.")
    assert.Equal(t, 0, len(deployment.Packages), "Get package list failed.")
    assert.Equal(t, "test_package", deployment.Package.Packagename, "Get package name failed.")
    assert.Equal(t, "/wskdeploy/samples/test", deployment.Package.Namespace, "Get package namespace failed.")
    assert.Equal(t, "12345678ABCDEF", deployment.Package.Credential, "Get package credential failed.")
    assert.Equal(t, 1, len(deployment.Package.Inputs), "Get package input list failed.")
    //get and verify inputs
    for param_name, param := range deployment.Package.Inputs {
        assert.Equal(t, "value", param.Value, "Get input value failed.")
        assert.Equal(t, "param", param_name, "Get input param name failed.")
    }
}

func TestParseDeploymentYAML_Action(t *testing.T) {
	mm := NewYAMLParser()
    deployment, _ := mm.ParseDeployment("../tests/dat/deployment_data_application_package.yaml")

	for pkg_name := range deployment.Application.Packages {

		var pkg = deployment.Application.Packages[pkg_name]
		for action_name := range pkg.Actions {
			assert.Equal(t, "hello", action_name, "Get action name failed.")
			var action = pkg.Actions[action_name]
			assert.Equal(t, "/wskdeploy/samples/test/hello", action.Namespace, "Get action namespace failed.")
			assert.Equal(t, "12345678ABCDEF", action.Credential, "Get action credential failed.")
			assert.Equal(t, 1, len(action.Inputs), "Get package input list failed.")
			//get and verify inputs
			for param_name, param := range action.Inputs {
				switch param.Value.(type) {
				case string:
					assert.Equal(t, "name", param_name, "Get input param name failed.")
					assert.Equal(t, "Bernie", param.Value, "Get input value failed.")
				default:
					t.Error("Get input value type failed.")
				}
			}
		}
	}
}
