# A Simple Local Password Manager: pmanager

## Features
- Ability to generate strong passwords.
- Ability to store passwords using AES encryption method.
- Ability to list and retrieve saved passwords.
- Works only with root permissions.
- Ecrypted passwords are located in /Library/pmanager directory with root owner. You can change the directory from source code.

## How to install
- Make sure to have Golang 1.19 installed.
- clone the repository.
- run ````go build -o pmanager main.go```` in root directory of the repository.
- sudo mv pmanager /usr/local/bin/

## How to use
- Run ````pmanager```` command from the terminal.
- There will be five options:
1. Generate Password
2. Retrieve Password
3. List Saved Websites
4. Delete Website
5. Exit