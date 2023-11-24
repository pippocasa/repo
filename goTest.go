package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"net/http"
	"io/ioutil" // Import the ioutil package
	"time" // Import the time package
	"flag"
	"os"
)
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
// Device structure to represent the JSON data
type DeviceInfo struct {
	ID       int64  `json:"id"`
	Hostname string `json:"hostname"`
	DeviceFunction string `json:"device_function"`
	IPAddress string `json:"ip_address"`
	SoftwareVersion string `json:"software_version"`
}

// Response structure for the JSON response
type Response struct {
	Page    int      `json:"page"`
	Count   int      `json:"count"`
	TotalPages int  `json:"total_pages"`
	TotalCount int  `json:"total_count"`
	Data    []DeviceInfo `json:"data"`
}
type Credentials struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	File     *string  `json:"file,omitempty"`
	Commands  []string `json:"commands"`
        Devices  *[]string `json:"devices,omitempty"` // List of device IP addresses (optional)
}

const baseURL = "https://api.extremecloudiq.com"
const version = "0.3" // Set the version number here
func readCredentials(filePath string) (Credentials, error) {
	var credentials Credentials

	// Read the credentials from the JSON file
	credentialsFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return credentials, fmt.Errorf("error reading credentials from file: %v", err)
	}

	// Unmarshal the JSON data into the Credentials struct
	if err := json.Unmarshal(credentialsFile, &credentials); err != nil {
		return credentials, fmt.Errorf("error decoding credentials from JSON: %v", err)
	}

	return credentials, nil
}


func login(credentials Credentials) (string, error) {
	// Prepare the JSON payload
	payload, err := json.Marshal(map[string]interface{}{
		"username": credentials.Username,
		"password": credentials.Password,
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON payload: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", baseURL+"/login", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Process the login response
	var loginResponse LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return "", fmt.Errorf("error decoding login response: %v", err)
	}

	return loginResponse.AccessToken, nil
}

func getDevices(token string, commands []string,deviceIPs *[]string) {
        var totalDevices []DeviceInfo
        var wg sync.WaitGroup
        devices, err := fetchDevices(token)
        if err != nil {
                fmt.Println("Error fetching devices:", err)
                return
        }
        if deviceIPs != nil {
		for _, device := range devices {
			for _, ip := range *deviceIPs {
				if device.IPAddress == ip {
					totalDevices = append(totalDevices, device)
					break
				}
			}
		}
	} else {
        // Filter devices where DeviceFunction is "AP"
        for _, device := range devices {
                if device.DeviceFunction == "AP" {
                        totalDevices = append(totalDevices, device)
                }
        }
        }
	fmt.Printf("Total Devices with DeviceFunction 'AP': %d\n", len(totalDevices))
	concurrencyFraction := 0.2 // You can adjust this value as needed
        concurrencyLimit := int(float64(len(totalDevices)) * concurrencyFraction)
        // Ensure that the concurrency limit is at least 1 to avoid issues
        if concurrencyLimit < 1 {
            concurrencyLimit = 1
        }
	// Print total devices count
	// Create a buffered channel for controlling concurrency
	ch := make(chan DeviceInfo, concurrencyLimit)
	// Start worker goroutines
	for i := 0; i < concurrencyLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for device := range ch {
			       fetchResult(device.ID, device.Hostname, token, commands)
			}
		}()
	}
	// Send devices to the channel
	for _, device := range totalDevices {
		ch <- device
	}

	// Close the channel to signal that no more items will be sent
	close(ch)

	// Wait for all worker goroutines to finish
	wg.Wait()
        // Return the captured output as a string
}

func fetchDevices(token string) ([]DeviceInfo, error) {
	baseURL := "https://api.extremecloudiq.com/devices"
        page := 1
        deviceType := "REAL"
        async := false
        var devices []DeviceInfo
	for {
		params := fmt.Sprintf("?page=%d&limit=%d&deviceTypes=%s&async=%v", page, 10, deviceType, async)
		requestURL := fmt.Sprintf("%s%s", baseURL, params)

		// Use the provided access token
		accessToken := token

		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+accessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
		}

		// Read the response body using io/ioutil
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		// Define a variable to store the extracted data
		var response Response

		// Unmarshal the JSON data from the response body into the Response struct
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("error unmarshaling JSON response: %v", err)
		}
		// Access the extracted values from the Data field
		devices = append(devices, response.Data...)
		if page >= response.TotalPages {
			break
		}
		page++
	}

	return devices, nil
}

func fetchResult(deviceID int64, hostname string, token string , commands[]string) {
    client := &http.Client{}
    payload, err := json.Marshal(commands)
	if err != nil {
		fmt.Println("Error marshaling JSON payload:", err)
		return
	}
    req, err := http.NewRequest("POST", fmt.Sprintf("https://api.extremecloudiq.com/devices/%d/:cli", deviceID), bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    //fmt.Println("Response Body:", string(bodyText))
    // Unmarshal the response into a struct
    var response struct {
        DeviceCLIOutputs map[int64][]struct {
            CLI          string `json:"cli"`
            ResponseCode string `json:"response_code"`
            Output       string `json:"output"`
        } `json:"device_cli_outputs"`
    }
    if err := json.Unmarshal(bodyText, &response); err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Check the response code and print the output if it's "SUCCEED"
    if deviceResponses, ok := response.DeviceCLIOutputs[deviceID]; ok {
        for _, entry := range deviceResponses {
            if entry.ResponseCode == "SUCCEED" {
                fmt.Printf("Device ID: %d, Hostname: %s, Output: %s\n", deviceID, hostname, entry.Output)
            }
        }
    }
}

func readConfig(configFile string) (*Credentials, error) {
	// Open and read the configuration file
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode JSON from the file into a Credentials struct
	var credentials Credentials
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&credentials); err != nil {
		return nil, err
	}

	return &credentials, nil
}


func writeToFile(filePath, content string) error {
	// Open the file for writing (create if not exists, append if exists)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write content to the file
	_, err = file.WriteString(content)
	return err
}


func main() {
        // Parse command-line flags
	configFile := flag.String("config", "credentials.json", "Path to the configuration file")
	helpFlag := flag.Bool("help", false, "Show help message")
	flag.Parse()

	// Check if the help flag is provided
	if *helpFlag {
		printHelp()
		return
	}
        fmt.Println("******************* @ExtremeNetworks 2023 *******************")
        fmt.Printf("Program Version: %s\n", version)
        fmt.Printf("\n")
	// Read the JSON configuration from the file
	credentials, err := readConfig(*configFile)
	if err != nil {
		fmt.Println("Error reading configuration:", err)
		return
	}

	// Use the values from the configuration
        if credentials.File != nil {
		// Create and write to the specified file
		if err := writeToFile(*credentials.File, "Output written to the file."); err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		fmt.Println("Output written to", *credentials.File)
	} else {
	//	fmt.Println("Output will be written to the console")
	}

	// Login process
	token, err := login(*credentials)
	if err != nil {
		fmt.Println("Login failed:", err)
		return
	}
        // Record the start time
        startTime := time.Now()
	fmt.Printf("Please wait, retrieving data %s\n",credentials.Commands)
	// Process the response body
	//getDevices(token, credentials.Commands)
        getDevices(token, credentials.Commands, credentials.Devices)
	// Record the end time
        endTime := time.Now()

        // Calculate the elapsed time
        elapsedTime := endTime.Sub(startTime)

        fmt.Printf("Data retrieval completed in %s\n", elapsedTime)
}

func printHelp() {
	// Print help message
	fmt.Println("Usage:")
	fmt.Println("  -config   Path to the configuration file (default: credentials.json)")
	fmt.Println("  -help     Show this help message")
}

