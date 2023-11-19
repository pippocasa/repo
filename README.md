# retrieveXIQData

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Overview

`retrieveXIQData` is a program that calls all devices from an XIQ account and sends a list of CLI commands to execute against all devices. The user name, password, and commands are configured in an external configuration file (`credentials.json` by default).

## Installation

### Windows
![Windows Icon](https://img.icons8.com/color/48/000000/windows-10.png)

1. Download `retrieveXIQData.exe`.
- You can download the executable from [here](https://github.com/pippocasa/repo/blob/main/retrieveXIQData.exe).
2. Create a `credentials.json` file with the following structure:

    ```json
    {
      "username": "your_username",
      "password": "your_password",
      "commands": [
        "command_1",
        "command_2",
        "command_3",
        // Add additional commands as needed.
        //Ensure that the last command in the "commands" array does not have a comma after it
      ]
    }
    ```
Note: If you have only one command in the "commands" array, you don't need a comma after it.

### MacOS

1. Download `retrieveXIQData.app`.
- You can download the executable from [here](https://github.com/pippocasa/repo/blob/main/retrieveXIQData.app).
2. Create a `credentials.json` file with the following structure:

    ```json
    {
      "username": "your_username",
      "password": "your_password",
      "commands": [
        "command_1",
        "command_2",
        "command_3",
        // Add additional commands as needed.
        //Ensure that the last command in the "commands" array does not have a comma after it
      ]
    }
    ```
Note: If you have only one command in the "commands" array, you don't need a comma after it.


### Manual Installation (for non-Windows systems)

```bash
# Clone the repository
git clone https://github.com/pippocasa/repo.git

# Change into the project directory
cd repo

# Build the program
go build
```
### Usage
### Windows 
![Windows Icon](https://img.icons8.com/color/48/000000/windows-10.png)

To use the program, on a console window launch the program with :

retrieveXIQData.exe

### MacOS
To use the program, on a console window launch the program with :

./retrieveXIQData.app

### Linux

./repo
### Help
For additional help and options, run the following command:

```bash
retrieveXIQData.exe -help
```

## Python version
If you want to use getDeviceList.py instead of the compiled executables retrieveXIQData.exe you need a python environment. The easiest way to check the Python version is to open the terminal and type this command 
```bash
python3 --version
```
credentials.json is also required and should be in the same directory as your script.

### Required Python modules
The requests module is required for the retrieveDeviceList.py script.You can check if the required modules are installed using the terminal
```bash
python3 -c "import requests" 
```
if a error message like 'a ModuleNotFoundError: No module named requests' is returned then you need to install the request module.

### Install required Python modules
The required modules can be installed using pip3 using the following command.
```bash
pip3 install requests
```
### Running the script
To run the script, open the terminal to the location of the script and run the following:
```bash
python3 getDeviceList.py
```
You can also make the script executable by running
```bash
chmod +x getDeviceList.py
```
Then, you can run the script by typing 
```bash
./getDeviceList.py
```

