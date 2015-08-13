package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/iron-io/iron_go3/api"
	"github.com/iron-io/iron_go3/worker"
)

// create code package (zip) from parsed .worker info
func pushCodes(zipName, host string, w *worker.Worker, args worker.Code) (id string, err error) {
	// TODO i don't get why i can't write from disk to wire, but I give up
	var body bytes.Buffer
	mWriter := multipart.NewWriter(&body)
	mMetaWriter, err := mWriter.CreateFormField("data")
	if err != nil {
		return "", err
	}
	reqMap := map[string]interface{}{
		"name":            args.Name,
		"config":          args.Config,
		"max_concurrency": args.MaxConcurrency,
		"retries":         args.Retries,
		"retries_delay":   args.RetriesDelay.Seconds(),
		"image":           args.Image,
	}
	if host != "" {
		reqMap["host"] = host
	}
	if args.Command != "" {
		reqMap["command"] = args.Command
	}

	jEncoder := json.NewEncoder(mMetaWriter)
	if err := jEncoder.Encode(reqMap); err != nil {
		return "", err
	}

	if zipName != "" {
		r, err := zip.OpenReader(zipName)
		if err != nil {
			return "", err
		}
		defer r.Close()

		mFileWriter, err := mWriter.CreateFormFile("file", "worker.zip")
		if err != nil {
			return "", err
		}
		zWriter := zip.NewWriter(mFileWriter)

		for _, f := range r.File {
			fWriter, err := zWriter.Create(f.Name)
			if err != nil {
				return "", err
			}
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			_, err = io.Copy(fWriter, rc)
			rc.Close()
			if err != nil {
				return "", err
			}
		}

		zWriter.Close()
	}
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

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if err = api.ResponseAsError(response); err != nil {
		return "", err
	}

	var data struct {
		Id string `json:"id"`
	}
	err = json.NewDecoder(response.Body).Decode(&data)
	return data.Id, err
}
