package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	tests := map[string]struct {
		yaml            string
		json            string
		mockFilePath    string
		mockFileContent []byte
		output          *fileData
		expectedError   bool
		errMessage      string
	}{
		"no file":            {yaml: "", json: "", output: nil, expectedError: true, errMessage: "must provide a file"},
		"two files provided": {yaml: "paths.yml", json: "paths.nil", output: nil, expectedError: true, errMessage: "must provide json or yaml but not both at the same time"},
		"expected json file": {yaml: "", json: "paths.yml", output: nil, expectedError: true, errMessage: "expected json file"},
		"expected yaml file": {yaml: "paths.json", json: "", output: nil, expectedError: true, errMessage: "expected yml or yaml file"},
		"reading yaml file": {
			yaml:         filepath.Join(dir, "paths.yml"),
			mockFilePath: filepath.Join(dir, "paths.yml"),
			mockFileContent: []byte(`
- path: /urlshort
	url: https://github.com/gophercises/urlshort
- path: /urlshort-final
	url: https://github.com/gophercises/urlshort/tree/solution
`),
			output: &fileData{
				data: []byte(`
- path: /urlshort
	url: https://github.com/gophercises/urlshort
- path: /urlshort-final
	url: https://github.com/gophercises/urlshort/tree/solution
`),
				isJSON: false,
				isYAML: true,
			},
			expectedError: false,
		},
		"reading json file": {
			json:         filepath.Join(dir, "paths.json"),
			mockFilePath: filepath.Join(dir, "paths.json"),
			mockFileContent: []byte(`
[
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
	{
		"path": "/urlshort-final",
		"url": "https://github.com/gophercises/urlshort/tree/solution"
	}
]
`),
			output: &fileData{
				data: []byte(`
[
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
	{
		"path": "/urlshort-final",
		"url": "https://github.com/gophercises/urlshort/tree/solution"
	}
]
`),
				isJSON: true,
				isYAML: false,
			},
			expectedError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// create mock file if needed
			if len(tc.mockFilePath) > 0 && len(tc.mockFileContent) > 0 {
				if err := os.WriteFile(tc.mockFilePath, tc.mockFileContent, 0644); err != nil {
					t.Fatalf("error creating mock file:: %v", err)
				}
			}

			output, err := readFile(&tc.yaml, &tc.json)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected: %v, got: %v", tc.expectedError, (err != nil))
			}
			if err != nil {
				if err.Error() != tc.errMessage {
					t.Fatalf("expected: %v, got: %v", tc.errMessage, err.Error())
				}
			}
			if !reflect.DeepEqual(output, tc.output) {
				t.Fatalf("expected: %+v, got: %+v", tc.output, output)
			}
		})
	}
}
