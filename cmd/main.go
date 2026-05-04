package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Fremenkiel/yogo-api-client/v2/internal/models"
	"github.com/Fremenkiel/yogo-api-client/v2/pkg/dotenv"
)

var (
	baseURL = "https://api.yogo.dk/"
)

type ClientAuth struct {
	Username		string
	Password		string
}

type YogoClient struct {
	client    *http.Client
	userAgent string
	jwt     string
	auth			ClientAuth
}

type YogoResponse []models.Customer
type AuthResponse struct {
	User	models.Customer		`json:"user"`
	Token	string	`json:"token"`
}

func NewYogoClient() *YogoClient {
	return &YogoClient{
		client:    &http.Client{Timeout: 30 * time.Second},
		userAgent: getRandomUserAgent(),
	}
}

func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents))))
	return userAgents[n.Int64()]
}

func (c *YogoClient) getToken() error {
	ctx := context.Background()
	fullURL := baseURL + "login"
	
    body := struct {
        Email string `json:"email"`
        Password string `json:"password"`
    }{
        Email: os.Getenv("USERNAME"),
        Password: os.Getenv("PASSWORD"),
    }

    out, err := json.Marshal(body)
    if err != nil {
        log.Fatal(err)
    }

    req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(out))
    if err != nil {
        log.Fatal(err)
    }

    req = req.WithContext(ctx)

	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    body, _ := io.ReadAll(resp.Body) // Read the error page
    return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var reader io.Reader = resp.Body
	resBody, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	var authResponse AuthResponse
	if err = json.Unmarshal(resBody, &authResponse); err != nil {
		return err
	}

	if authResponse.Token == "" {
		return errors.New("No token rec")
	}
	log.Print(authResponse.Token)

	c.jwt = authResponse.Token
	return nil
}

func (vc *YogoClient) makeRequest(reqType string, endpoint string, params map[string]string) ([]byte, error) {
	fullURL := baseURL + endpoint

	if len(params) > 0 {
		u, _ := url.Parse(fullURL)
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	req, err := http.NewRequest(reqType, fullURL, nil)
	if err != nil {
		return nil, err
	}

	if vc.jwt == "" {
		log.Print("Getting token")
		err = vc.getToken()
		if err != nil {
			log.Fatal(err)
		}
	}

	vc.setHeaders(req)

	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    body, _ := io.ReadAll(resp.Body) // Read the error page
    return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var reader io.Reader = resp.Body

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *YogoClient) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Referer",	os.Getenv("CUSTOMER_URL"))
	req.Header.Set("Origin", os.Getenv("CUSTOMER_URL"))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Yogo-Request-Context", "admin")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.jwt))
}

func (c *YogoClient) GetCustomers() ([]models.Customer, error) {
	var yogoResponse YogoResponse
	endpoint := "users"
	response, err := c.makeRequest("GET", endpoint, map[string]string{
		"customer": "true",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Raw response length: %d", len(response))
	log.Printf("Raw response content: %s", string(response)) // Look at this!

	if len(response) == 0 {
		return nil, fmt.Errorf("received empty response from server")
	}

	if err = json.Unmarshal(response, &yogoResponse); err != nil {
		return nil, err
	}
	return yogoResponse, err
}

func main() {
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		if err := dotenv.Load(envFile); err != nil {
			log.Fatalf("Unable to load env file %s: %v", envFile, err)
		}
	}
	client := NewYogoClient()
	customers, err := client.GetCustomers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	log.Print(customers)
}
