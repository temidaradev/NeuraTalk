package ollama

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func QueryOllamaModel(input string) (string, error) {
	url := "http://localhost:5000/query" // Replace with the actual URL of your Ollama model

	requestBody, err := json.Marshal(map[string]string{
		"input": input,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result["output"], nil
}
