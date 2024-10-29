package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const storageDir = "/Library/pmanager"
const defaultPasswordLength = 10
const encryptionKey = "some-secret-key" // replace with a secure key

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Please run this program as root or with sudo.")
		os.Exit(1)
	}

	for {
		fmt.Println("\nChoose an option:")
		fmt.Println("1. Generate Password")
		fmt.Println("2. Retrieve Password")
		fmt.Println("3. List Saved Websites")
		fmt.Println("4. Delete Website")
		fmt.Println("5. Exit") // Added exit option

		var option int
		fmt.Scan(&option)

		switch option {
		case 1:
			generatePassword()
		case 2:
			retrievePassword()
		case 3:
			listWebsites()
		case 4:
			deleteWebsite()
		case 5:
			fmt.Println("Exiting the program. Goodbye!")
			os.Exit(0) // Gracefully exit the program
		default:
			fmt.Println("Invalid option selected.")
		}
	}
}

// Deletes the saved file for the specified website
func deleteWebsite() {
	fmt.Print("Enter website name to delete: ")
	var website string
	fmt.Scan(&website)

	// Check if website name is empty
	if website == "" {
		fmt.Println("Website name cannot be empty. Please enter a valid name.")
		return
	}

	filePath := storageDir + "/" + website

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Website '%s' does not exist.\n", website)
		return
	}

	// Remove the file
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("Error deleting website:", err)
		return
	}

	fmt.Printf("Website '%s' deleted successfully.\n", website)
}

// Generates a secure random password and optionally saves it
func generatePassword() {
	fmt.Print("Enter desired password length (default 10): ")
	var length int
	_, err := fmt.Scanf("%d", &length)

	// Check if input is non-numeric or empty, or if it exceeds the max length
	if err != nil || length > 32 {
		if err != nil {
			fmt.Println("Invalid input. Please enter a numeric value.")
		} else {
			fmt.Println("Password length cannot exceed 32. Using default length of 10.")
		}
		length = defaultPasswordLength
	} else if length == 0 {
		length = defaultPasswordLength // Use default if user presses Enter without input
	}

	password, err := generateRandomPassword(length)
	if err != nil {
		fmt.Println("Error generating password:", err)
		return
	}
	fmt.Printf("Generated Password: %s\n", password)

	fmt.Print("Do you want to save this password? (y/n): ")
	var saveOption string
	fmt.Scan(&saveOption)
	if strings.ToLower(saveOption) == "y" {
		fmt.Print("Enter website name: ")
		var website string
		fmt.Scan(&website)
		savePassword(website, password)
		fmt.Println("Password saved successfully.")
	}
}

// Generates a random password with specified length
func generateRandomPassword(length int) (string, error) {
	// Ensure minimum length of 8 characters
	if length < 8 {
		length = 8
	}

	const (
		lowerCharset   = "abcdefghijklmnopqrstuvwxyz"
		upperCharset   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digitCharset   = "0123456789"
		specialCharset = "!@#$%^&*()-_=+"
	)

	// Ensure at least one character from each required category
	requiredChars := []byte{
		upperCharset[getRandomIndex(len(upperCharset))],
		digitCharset[getRandomIndex(len(digitCharset))],
		specialCharset[getRandomIndex(len(specialCharset))],
	}

	// Fill remaining password characters randomly from all sets
	allCharset := lowerCharset + upperCharset + digitCharset + specialCharset
	password := make([]byte, length)
	copy(password, requiredChars)

	for i := len(requiredChars); i < length; i++ {
		password[i] = allCharset[getRandomIndex(len(allCharset))]
	}

	// Shuffle password to randomize the order of characters
	shuffledPassword := shuffle(password)

	return string(shuffledPassword), nil
}

// Generates a random index within the given limit using crypto/rand
func getRandomIndex(limit int) int {
	index := make([]byte, 1)
	_, err := rand.Read(index)
	if err != nil {
		panic(err) // Handle this appropriately in production code
	}
	return int(index[0]) % limit
}

// Shuffles the byte slice for randomness
func shuffle(slice []byte) []byte {
	shuffled := make([]byte, len(slice))
	copy(shuffled, slice)
	for i := range shuffled {
		j := getRandomIndex(len(shuffled))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

// Encrypts and saves the password for the given website
func savePassword(website, password string) {
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		fmt.Println("Error creating storage directory:", err)
		os.Exit(1)
	}

	encryptedPassword, err := encrypt(password, encryptionKey)
	if err != nil {
		fmt.Println("Error encrypting password:", err)
		os.Exit(1)
	}

	filePath := storageDir + "/" + website
	if err := ioutil.WriteFile(filePath, []byte(encryptedPassword), 0600); err != nil {
		fmt.Println("Error saving password:", err)
		os.Exit(1)
	}
}

// Retrieves and decrypts the password for a given website
func retrievePassword() {
	fmt.Print("Enter website name to retrieve password: ")
	var website string
	fmt.Scan(&website)

	// Check if website name is empty
	if website == "" {
		fmt.Println("Website name cannot be empty. Please enter a valid name.")
		return
	}

	filePath := storageDir + "/" + website

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Password for website '%s' does not exist.\n", website)
		return
	}

	// Read and decrypt the password
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error retrieving password:", err)
		os.Exit(1)
	}

	password, err := decrypt(string(data), encryptionKey)
	if err != nil {
		fmt.Println("Error decrypting password:", err)
		os.Exit(1)
	}

	fmt.Printf("Password for %s: %s\n", website, password)
}

// Lists all saved websites
func listWebsites() {
	files, err := ioutil.ReadDir(storageDir)
	if err != nil {
		fmt.Println("Error listing websites:", err)
		os.Exit(1)
	}

	fmt.Println("Saved websites:")
	for _, file := range files {
		if !file.IsDir() {
			fmt.Println("- " + file.Name())
		}
	}
}

// Encrypts a plain text password with AES
func encrypt(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(createHash(key)))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// Decrypts the AES-encrypted password
func decrypt(cipherText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(createHash(key)))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.URLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, cipherTextBytes := data[:nonceSize], data[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherTextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// Creates a hash of the key for AES encryption
func createHash(key string) string {
	hash := md5.Sum([]byte(key))
	return string(hash[:])
}
