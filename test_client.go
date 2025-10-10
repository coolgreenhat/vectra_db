package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Vector struct {
	ID       string            `json:"id"`
	Vector   []float64         `json:"vector"`
	Metadata map[string]string `json:"metadata"`
}

type SearchParams struct {
	Query  []float64          `json:"query"`
	Filter map[string]string  `json:"filter"`
	Page   int                `json:"page"`
	Limit  int                `json:"limit"`
}

func main() {
	baseURL := "http://localhost:8081"

	// ---------- [POST] /vectors ----------
	vectors := []Vector{
		{ID: "v101", Vector: []float64{0.3, 0.6, 0.1, 0.8}, Metadata: map[string]string{"author": "Tara", "topic": "AI"}},
		{ID: "v102", Vector: []float64{0.9, 0.2, 0.5, 0.1}, Metadata: map[string]string{"author": "Raj", "topic": "Math"}},
		{ID: "v103", Vector: []float64{0.4, 0.7, 0.3, 0.2}, Metadata: map[string]string{"author": "Nina", "topic": "AI"}},
	}

	for _, v := range vectors {
		resp, body := post(baseURL+"/vectors", v)
		expectStatus("POST /vectors", resp, body, http.StatusCreated)
	}

	// ---------- [GET] /vectors/v102 ----------
	resp, body := get(baseURL + "/vectors/v102")
	expectStatus("GET /vectors/v102", resp, body, http.StatusOK)
	expectBodyMatch("GET /vectors/v102", resp, body, vectors[1])

	// ---------- [POST] /search ----------
	search := SearchParams{
		Query:  []float64{0.4, 0.7, 0.3, 0.2},
		Filter: map[string]string{"topic": "AI"},
		Page:   1,
		Limit:  10,
	}
	resp, body = post(baseURL+"/search", search)
	expectStatus("POST /search", resp, body, http.StatusOK)

	// ---------- [PUT] /vectors/v103 (nonexistent path test) ----------
	resp, body = put(baseURL+"/Vectors/v103", map[string]any{"vector": []float64{0.2, 0.9, 0.1, 0.4}})
	expectStatus("PUT /Vectors/v103", resp, body, http.StatusNotFound)

	// ---------- [DELETE] /vectors/v101 ----------
	resp, body = deleteReq(baseURL + "/vectors/v101")
	expectStatus("DELETE /vectors/v101", resp, body, http.StatusNoContent)

	// ---------- [GET] /vectors/v101 (should be deleted) ----------
	resp, body = get(baseURL + "/vectors/v101")
	expectStatus("GET /vectors/v101", resp, body, http.StatusNotFound)

	fmt.Println("\n✅ All tests executed.")
}

// -------------------- Helpers --------------------

func post(url string, data interface{}) (*http.Response, []byte) {
	return doRequest("POST", url, data)
}

func put(url string, data interface{}) (*http.Response, []byte) {
	return doRequest("PUT", url, data)
}

func deleteReq(url string) (*http.Response, []byte) {
	return doRequest("DELETE", url, nil)
}

func get(url string) (*http.Response, []byte) {
	return doRequest("GET", url, nil)
}

func doRequest(method, url string, data interface{}) (*http.Response, []byte) {
	var body io.Reader
	if data != nil {
		b, _ := json.Marshal(data)
		body = bytes.NewBuffer(b)
	}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return nil, nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

func expectStatus(name string, resp *http.Response, body []byte, want int) {
	if resp == nil {
		fmt.Printf("❌ %s: no response\n", name)
		return
	}
	if resp.StatusCode != want {
		fmt.Printf("❌ %s: expected %d, got %d\nBody: %s\n", name, want, resp.StatusCode, string(body))
	} else {
		fmt.Printf("✅ %s: %d OK\n", name, resp.StatusCode)
	}
}

func expectBodyMatch(name string, resp *http.Response, body []byte, want Vector) {
	if resp == nil {
		fmt.Printf("❌ %s: no response\n", name)
		return
	}
	var got Vector
	_ = json.Unmarshal(body, &got)
	if got.ID != want.ID {
		fmt.Printf("❌ %s: expected %v, got %v\n", name, want, got)
	} else {
		fmt.Printf("✅ %s: body OK\n", name)
	}
}

