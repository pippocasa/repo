"""
******************************************************************************
 * DISCLAIMER: 
 * 
 * This code was written by Giuseppe Casablanca gcasabla@extremenetworks.com and is provided "as is" without 
 * warranty of any kind, either express or implied, including, but not limited 
 * to, the implied warranties of merchantability and fitness for a particular 
 * purpose. 
 * 
 * Giuseppe Casablanca gcasabla@extremenetworks.com and any contributors to this code make no claims or guarantees 
 * regarding its functionality, correctness, or performance. You are using 
 * this code at your own risk.
 * 
***************************************************************************
"""

import requests
import json
import re
import threading
from concurrent.futures import ThreadPoolExecutor
import logging
from datetime import datetime
# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger()

def read_credentials_from_file(file_path):
    try:
        with open(file_path, 'r') as file:
            credentials = json.load(file)
            return credentials.get("username"), credentials.get("password"),credentials.get("commands",[])
    except FileNotFoundError:
        logger.error(f"Credentials file not found: {file_path}")
        return None, None,[]

def login(username, password):
    login_url = "https://api.extremecloudiq.com/login"
    login_data = {
        "username": username,
        "password": password
    }

    try:
        login_response = requests.post(login_url, json=login_data)
        login_response.raise_for_status()  # Raise an exception for non-200 status codes
        token = login_response.json().get("access_token")
        return token
    except requests.exceptions.RequestException as e:
        logger.error(f"Failed to log in: {e}")
        return None

def fetch_device_page(page, headers):
    device_list_url = "https://api.extremecloudiq.com/devices"
    per_page = 10
    params = {"page": page, "limit": per_page, "deviceTypes": "REAL", "async": "false"}

    response = requests.get(device_list_url, headers=headers, params=params)
    if response.status_code == 200:
        devices = response.json()["data"]
        return devices
    else:
        print(f"Failed to retrieve devices for page {page}")
        return []

def fetch_device_list(token, headers):
    total_devices = []
    page = 1

    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = []
        while True:
            futures.append(executor.submit(fetch_device_page, page, headers))
            page += 1

            if page > 30:  # Set a reasonable limit for the number of pages to fetch
                break

        for future in futures:
            devices = future.result()
            if devices:
                total_devices.extend(devices)

    return total_devices

# Function to fetch speed information for a device
def fetch_speed(device_id, hostname, headers,commands):
    test=[] 
    device_commands_url = f"https://api.extremecloudiq.com/devices/{device_id}/:cli"
     #command = ['show interface eth0 | incl speed']
    test.append(commands)
    command=test
    #command = commands
    # Convert the list to JSON
    #data = json.dumps(commands)  
    device_command_response = requests.post(device_commands_url, headers=headers, json=command)
    if device_command_response.status_code == 200:
        # Parse the JSON data
        data = json.loads(device_command_response.text)
        for response_device_id, responses in data["device_cli_outputs"].items():
            response_code = responses[0]["response_code"]
            if response_code == "SUCCEED":
                speed_info = responses[0]["output"]
                speed = speed_info.split("Speed=")[1].split(";")[0]
                print(f"Device ID: {response_device_id}, Name: {hostname}, Speed: {speed}")
            else:
                print(f"Device ID: {response_device_id}, Response code is not 'SUCCEED'. No speed information extracted.")


def fetch_result(device_id, hostname, headers,commands):
    device_commands_url = f"https://api.extremecloudiq.com/devices/{device_id}/:cli"
    #command = ['show interface eth0 | incl speed']
    device_command_response = requests.post(device_commands_url, headers=headers, json=commands)

    if device_command_response.status_code == 200:
        # Parse the JSON data
        data = json.loads(device_command_response.text)
        # Access the data
        for response_device_id, responses in data["device_cli_outputs"].items():
            for entry in responses:
                 response_code = entry["response_code"]
                 if response_code == "SUCCEED":
                     output = entry["output"]
                     print(f"Device ID: {response_device_id}, Name: {hostname}, Output: {output}")
            print("\n")  # Separate the output of different devices 

def main():
    # Configuration settings
    # Read credentials from a file
    username, password,commands = read_credentials_from_file('credentials.json')
    #commands=command.split(",")
    if username is None or password is None:
        logger.error("Failed to read credentials from the file.")
        return

    # Step 1: Login and obtain an authentication token
    token = login(username, password)

    headers = {
        'Content-Type': 'application/json',
        "Authorization": f"Bearer {token}"
    }

    # Fetch device list
    start_time = datetime.now()
    print("******************* @ExtremeNetworks 2023 *******************")
    print(f"Please wait, retrieving data {commands}\n")
    device_list = fetch_device_list(token,headers)
    print(f"Total Devices: {len(device_list)}") 
    threads = []
    for device_data in device_list:
        device_id = device_data['id']
        hostname = device_data['hostname']
        t = threading.Thread(target=fetch_result, args=(device_id, hostname, headers, commands))
        threads.append(t)
        t.start()
    # Wait for all threads to finish
    for t in threads:
        t.join()
    end_time = datetime.now()

    # Calculate the duration
    duration = end_time - start_time
    print("Data retrieval completed in {}".format(duration))
if __name__== "__main__":
    main()
