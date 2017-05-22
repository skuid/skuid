package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"os"

	"bytes"

	"path/filepath"
)

type PullResponse struct { 
	Name               string  `json:"name"`
	UniqueID           string  `json:"uniqueId"`
	Type               string  `json:"type"`
	Module             string  `json:"module"`
	MaxAutoSaves       int     `json:"maxAutoSaves"`
	MasterPageUniqueID string  `json:"masterPageUniqueId,omitempty"`
	IsMasterPage       bool    `json:"isMasterPage"`
	ComposerSettings   *string `json:"composerSettings"`
	Body               string  `json:"body,omitempty"`
}

type PagePost struct {
	Changes   []PullResponse `json:"changes"`
	Deletions []PullResponse `json:"deletions"`
}

type PagePostResult struct {
	OrgName string   `json:"orgName"`
	Success bool     `json:"success"`
	Errors  []string `json:"upsertErrors,omitempty"`
}

type RetrieveRequest struct {
	Metadata        RetrieveMetadata `json:metadata`
}

type RetrieveMetadata struct {
	Apps            map[string]string `json:"apps,omitempty"`
    DataSources     map[string]string `json:"dataSources,omitempty"`
    Pages           map[string]string `json:"pages,omitempty"`
    Profiles        map[string]string `json:"profiles,omitempty"`
    Themes          map[string]string `json:"themes,omitempty"`
}

func (page *PullResponse) FileBasename() string {

	var buf bytes.Buffer

	if page.Module != "" {
		buf.WriteString(page.Module)
		buf.WriteString("_")
	}

	buf.WriteString(page.Name)

	return buf.String()
}

func (page *PullResponse) WriteAtRest(path string) (err error) {
	//if the desired directory isn't there, create it
	if _, err := os.Stat(path); err != nil {
		os.Mkdir(path, 0700)
	}
	//create a copy of the page to keep the Body out of the json
	clone := *page
	clone.Body = ""
	str, _ := json.MarshalIndent(clone, "", "    ")
	//write the json metadata
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.json", path, page.FileBasename()), str, 0644)

	if err != nil {
		return err
	}
	xml := page.Body
	//write the body to the file
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.xml", path, page.FileBasename()), []byte(xml), 0644)

	if err != nil {
		return err
	}

	return nil

}

func FilterByGlob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func FilterByModule(dir, moduleFilter string) ([]string, error) {

	filter := &bytes.Buffer{}

	filter.WriteString(moduleFilter)
	filter.WriteString("_*")

	pattern := filepath.Join(dir, filter.String())
	return filepath.Glob(pattern)
}

func filterOutXmlFiles(files []string) []string {
	filtered := []string{}

	for _, path := range files {
		if filepath.Ext(path) == ".xml" {
			continue
		}

		filtered = append(filtered, path)
	}

	return filtered
}

func ReadFiles(dir, moduleFilter, file string) ([]PullResponse, error) {

	var files []string

	if file != "" {
		var err error
		files, err = FilterByGlob(file)

		if err != nil {
			return nil, err
		}

	} else {
		var err error
		if _, err := os.Stat(dir); err != nil {
			return nil, err
		}

		files, err = FilterByModule(dir, moduleFilter)

		if err != nil {
			return nil, err
		}
	}

	files = filterOutXmlFiles(files)

	definitions := []PullResponse{}

	for _, file := range files {

		metadataFilename := file

		bodyFilename := strings.Replace(file, ".json", ".xml", 1)
		//read the metadata file
		metadataFile, _ := ioutil.ReadFile(metadataFilename)
		//read the page xml
		bodyFile, _ := ioutil.ReadFile(bodyFilename)

		pullRes := &PullResponse{}

		_ = json.Unmarshal(metadataFile, pullRes)

		pullRes.Body = string(bodyFile)

		definitions = append(definitions, *pullRes)
	}

	return definitions, nil
}
