package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/iron-io/iron_go/api"
	"github.com/iron-io/iron_go/worker"
	"gopkg.in/inconshreveable/log15.v2"
)

// create code package (zip) from parsed .worker info
func pushCodes(zipName, command string, w *worker.Worker, args worker.Code) (id string, err error) {
	r, err := zip.OpenReader(zipName)
	if err != nil {
		return "", err
	}
	defer r.Close()

	// TODO i don't get why i can't write from disk to wire, but I give up
	var body bytes.Buffer
	mWriter := multipart.NewWriter(&body)
	mMetaWriter, err := mWriter.CreateFormField("data")
	if err != nil {
		return "", err
	}
	jEncoder := json.NewEncoder(mMetaWriter)
	err = jEncoder.Encode(map[string]interface{}{
		"name":            args.Name,
		"command":         command,
		"config":          args.Config,
		"max_concurrency": args.MaxConcurrency,
		"retries":         args.Retries,
		"retries_delay":   args.RetriesDelay.Seconds(),
		"stack":           args.Stack,
		"image":           args.Image,
	})
	if err != nil {
		return "", err
	}
	mFileWriter, err := mWriter.CreateFormFile("file", "worker.zip")
	if err != nil {
		return "", err
	}
	zWriter := zip.NewWriter(mFileWriter)

	for _, f := range r.File {
		fWriter, err := zWriter.Create(f.Name)
		if err != nil {
			log15.Info("hi2")
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			log15.Info("hi3")
			return "", err
		}
		_, err = io.Copy(fWriter, rc)
		rc.Close()
		if err != nil {
			log15.Info("hi4")
			return "", err
		}
	}

	zWriter.Close()
	mWriter.Close()

	req, err := http.NewRequest("POST", api.Action(w.Settings, "codes").URL.String(), &body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("Authorization", "OAuth "+w.Settings.Token)
	req.Header.Set("Content-Type", mWriter.FormDataContentType())
	req.Header.Set("User-Agent", w.Settings.UserAgent)

	// dumpRequest(req) NOTE: never do this here, it breaks stuff
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if err = api.ResponseAsError(response); err != nil {
		return "", err
	}

	// dumpResponse(response)

	var data struct {
		Id         string `json:"id"`
		Msg        string `json:"msg"`
		StatusCode int    `json:"status_code"`
	}
	err = json.NewDecoder(response.Body).Decode(&data)
	return data.Id, err
}
