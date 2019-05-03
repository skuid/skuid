package cmd

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/skuid/skuid/types"
	"github.com/stretchr/testify/assert"
)

type RetrieveFile struct {
	Name string
	Body string
}

func TestRetrieve(t *testing.T) {
	testCases := []struct {
		testDescription string
		giveFiles       []RetrieveFile
		wantFiles       []RetrieveFile
		wantDirectories []string
		wantError       error
	}{
		{
			"retrieve nothing",
			[]RetrieveFile{},
			[]RetrieveFile{},
			[]string{},
			nil,
		},
		{
			"retrieve nonvalid skuid metadata files",
			[]RetrieveFile{
				{"readme.txt", "This archive contains some text files."},
				{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
				{"todo.txt", "Get animal handling licence.\nWrite more examples."},
			},
			[]RetrieveFile{},
			[]string{},
			nil,
		},
		{
			"retrieve a data source",
			[]RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
			},
			[]RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
			},
			[]string{
				"datasources",
			},
			nil,
		},
		{
			"retrieve two data sources",
			[]RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
				{"datasources/mydatasource2.json", "this is not even close to good JSON2"},
			},
			[]RetrieveFile{
				{"datasources/mydatasource.json", "this is not even close to good JSON"},
				{"datasources/mydatasource2.json", "this is not even close to good JSON2"},
			},
			[]string{
				"datasources",
			},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testDescription, func(t *testing.T) {
			// Create a buffer to write our archive to.
			buf := new(bytes.Buffer)

			// Create a new zip archive.
			w := zip.NewWriter(buf)

			for _, file := range tc.giveFiles {
				f, err := w.Create(file.Name)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.Write([]byte(file.Body))
				if err != nil {
					t.Fatal(err)
				}
			}

			// Make sure to check the error on Close.
			err := w.Close()
			if err != nil {
				t.Fatal(err)
			}

			filesCreated := []RetrieveFile{}
			directoriesCreated := []string{}

			mockFileMaker := func(fileReader io.ReadCloser, path string) error {
				body, err := ioutil.ReadAll(fileReader)
				if err != nil {
					t.Fatal(err)
				}
				filesCreated = append(filesCreated, RetrieveFile{
					Name: path,
					Body: string(body),
				})
				return nil
			}

			mockDirectoryMaker := func(path string, fileMode os.FileMode) error {
				if !types.StringSliceContainsKey(directoriesCreated, path) {
					directoriesCreated = append(directoriesCreated, path)
				}
				return nil
			}

			bufData := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

			err = writeResultsToDisk([]*io.ReadCloser{&bufData}, mockFileMaker, mockDirectoryMaker)
			assert.Equal(t, tc.wantFiles, filesCreated)
			assert.Equal(t, tc.wantDirectories, directoriesCreated)
			if tc.wantError != nil {
				assert.Equal(t, tc.wantError, err)
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
