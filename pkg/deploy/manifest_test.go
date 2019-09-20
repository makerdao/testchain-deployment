package deploy

import (
	"fmt"
	"testing"
)

func TestReadManifestFile(t *testing.T) {
	// Mock
	stepConfigTmpl := `{
		"description": "%s",
		"defaults": {},
		"roles": ["CREATOR"],
		"omniaFromAddr": "0xdc9A20F5a46AFE0802b361076BeFC51f787B2e58",
		"pauseDelay": "0"
	}`
	mockFileReader := func(path string) ([]byte, error) {
		t.Logf("mock file reader: %s", path)
		switch path {
		case ".staxx-scenarios":
			return []byte(`{
				"name": "TestManifest",
				"description": "",
				"scenarios": [
					{
						"name": "TestScenario1",
						"description": "",
						"run": "deploy-step-1",
						"config": "testconfig1.json"
					},
					{
						"name": "TestScenario2",
						"description": "",
						"run": "deploy-step-2",
						"config": "testconfig2.json"
					}
				]
			}`), nil
		default:
			return []byte(fmt.Sprintf(stepConfigTmpl, path)), nil
		}
	}

	// Run
	manifest, err := readManifestFile(mockFileReader, ".")
	if err != nil {
		t.Error(err)
	}

	scenario1 := manifest.Scenarios[0]
	scenario2 := manifest.Scenarios[1]

	// Assertion
	if manifest.Name != "TestManifest" {
		t.Errorf("Manifest name doesn't match unmarshaled name: %s", manifest.Name)
	}
	if scenario1.Name != "TestScenario1" {
		t.Errorf("Scenario 1's name doesnt match unmarshaled name: %s", scenario1.Name)
	}
	if scenario1.RunCommand != "deploy-step-1" {
		t.Errorf("Scenario 1's run command doesn't match unmarshaled run command: %s", scenario1.RunCommand)
	}
	if string(scenario1.Config) != fmt.Sprintf(stepConfigTmpl, "testconfig1.json") {
		t.Errorf("Scenario 1's config file content doesn't match: %s", scenario1.Config)
	}
	if scenario2.Name != "TestScenario2" {
		t.Errorf("Scenario 2's name doesn't match unmarshaled name: %s", scenario2.Name)
	}
}

func TestNewStepListFromManifest(t *testing.T) {
	// Setup
	manifest := Manifest{
		"TestManifest",
		"A test manifest",
		[]Scenario{
			{
				"TestScenario1!",
				"A test scenario",
				"true",
				[]byte(
					`{
						"description": "Step 1",
						"defaults": {},
						"roles": ["CREATOR"],
						"omniaFromAddr": "0xdc9A20F5a46AFE0802b361076BeFC51f787B2e58",
						"pauseDelay": "0"
					}`,
				),
			},
			{
				"TestScenario2!",
				"Another test scenario",
				"true",
				[]byte(
					`{
						"description": "Step 2",
						"defaults": {},
						"roles": ["CREATOR"],
						"omniaFromAddr": "0xdc9A20F5a46AFE0802b361076BeFC51f787B2e58",
						"pauseDelay": "0"
					}`,
				),
			},
		},
	}

	// Run
	stepList, err := NewStepListFromManifest(&manifest)
	if err != nil {
		t.Error(err)
	}

	step1 := stepList[0]
	step2 := stepList[1]

	// Assertion
	if step1.ID != 1 {
		t.Errorf("Scenario nr 1 doesn't map to step id 1, but instead: %d", step1.ID)
	}
	if step1.Description != "Step 1" {
		t.Errorf("Scenario nr 1's description doesn't match step 1's: %s", step1.Description)
	}
	if step1.OmniaFromAddress != "0xdc9A20F5a46AFE0802b361076BeFC51f787B2e58" {
		t.Errorf("Scenario nr 1's config file content doesn't match step 1's defaults: %s", step1.Defaults)
	}
	if string(step1.Defaults) != "{}" {
		t.Errorf("Scenario nr 1's config file content doesn't match step 1's defaults: %s", step1.Defaults)
	}

	if step2.ID != 2 {
		t.Errorf("Scenario nr 2 doesn't map to step id 2, but instead: %d", step2.ID)
	}
	if step2.Description != "Step 2" {
		t.Errorf("Scenario nr 2's description doesn't match step 2's: %s", step2.Description)
	}
	if string(step2.Defaults) != "{}" {
		t.Errorf("Scenario nr 2's config file content doesn't match step 2's defaults: %s", step2.Defaults)
	}
}
