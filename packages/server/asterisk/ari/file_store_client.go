package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	model_file "github.com/openline-ai/openline-customer-os/packages/server/file-store-api/model"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type FileStoreClient struct {
	username string
	tenant   string
	config   *RecordServiceConfig
}

func NewFileStoreClient(username string, tenant string, config *RecordServiceConfig) *FileStoreClient {
	return &FileStoreClient{
		username: username,
		tenant:   tenant,
		config:   config,
	}
}

func (fsc *FileStoreClient) UploadFile(filename string) (string, error) {
	file, _ := os.Open(filename)
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("audio", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()
	r, _ := http.NewRequest("POST", fmt.Sprintf("%s/file", fsc.config.FileStoreApiService), body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	r.Header.Add("Accept", "application/json")
	r.Header.Add("X-Openline-API-KEY", fsc.config.FileStoreApiKey)
	r.Header.Add("X-Openline-USERNAME", fsc.username)
	r.Header.Add("X-Openline-TENANT", fsc.tenant)

	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		log.Printf("UploadFile: could not send request: %s\n", err)
		return "", err
	}

	log.Printf("UploadFile: got response!\n")
	log.Printf("TranscribeAudio: status code: %d\n", res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("UploadFile: could not read response body: %s\n", err)
		return "", err

	}

	fileInfo := &model_file.File{}
	err = json.Unmarshal(resBody, fileInfo)
	if err != nil {
		log.Printf("UploadFile: could not unmarshal response body: %s\n", err)
		return "", err
	}
	return fileInfo.ID, nil
}
