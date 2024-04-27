package controller

import (
	"backend/config"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	mathRand "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// Global variables in controller package goes here
var timeoutLimit = 30 * time.Second
var validate = validator.New()

var accountCollection = config.GetCollection(config.ConnectDB(), "accounts")
var authorizationCollection = config.GetCollection(config.ConnectDB(), "authorizations")
var employeeCollection = config.GetCollection(config.ConnectDB(), "employee")
var epicCollection = config.GetCollection(config.ConnectDB(), "epics")
var messageCollection = config.GetCollection(config.ConnectDB(), "messages")
var projectCollection = config.GetCollection(config.ConnectDB(), "projects")
var taskCollection = config.GetCollection(config.ConnectDB(), "tasks")
var userInforCollection = config.GetCollection(config.ConnectDB(), "user_infor")

const (
	letterBytes  = "abcdefghijklmnopqrstuvwxyz"
	upperBytes   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialBytes = "!@#$%^&*()_+-=[]{};:,.<>/|?`~"
	numBytes     = "0123456789"
)

func GenerateAndHashPassword() (string, string) {
	randomGenerator := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	passwordLength, _ := strconv.Atoi(os.Getenv("PASSWORD_LENGTH"))
	password := ""
	for {
		password = ""
		for i := 0; i < passwordLength; i++ {
			randNum := randomGenerator.Intn(4)

			switch randNum {
			case 0:
				password += string(letterBytes[randomGenerator.Intn(len(letterBytes))])
			case 1:
				password += string(upperBytes[randomGenerator.Intn(len(upperBytes))])
			case 2:
				password += string(specialBytes[randomGenerator.Intn(len(specialBytes))])
			case 3:
				password += string(numBytes[randomGenerator.Intn(len(numBytes))])
			}
		}

		// Check if password is good
		if CheckPassword(password) {
			break
		}
	}

	combined := os.Getenv("PEPPER1") + password + os.Getenv("PEPPER2")

	// Hash the generated password
	hashedPassword, errhashedPassword := bcrypt.GenerateFromPassword([]byte(combined), 15)
	if errhashedPassword != nil {
		return "", password
	}

	hashedPasswordHex := hex.EncodeToString(hashedPassword)

	return hashedPasswordHex, password
}

func VerifyPassword(password, hashedPasswordHex string) bool {
	hashedPasswordBytes, decodeErr := hex.DecodeString(hashedPasswordHex)
	if decodeErr != nil {
		return false
	}

	combined := os.Getenv("PEPPER1") + password + os.Getenv("PEPPER2")

	// Verify the password
	errComparingPassword := bcrypt.CompareHashAndPassword(hashedPasswordBytes, []byte(combined))
	return errComparingPassword == nil
}

func CheckPassword(password string) bool {
	passwordLength, _ := strconv.Atoi(os.Getenv("PASSWORD_LENGTH"))

	if len(password) < passwordLength {
		return false
	}

	if os.Getenv("USE_SPECIAL") == "true" {
		valid := false
		for _, char := range specialBytes {
			if strings.Contains(password, string(char)) {
				valid = true
				break
			}
		}

		if !valid {
			return false
		}
	}

	if os.Getenv("USE_UPPER") == "true" {
		valid := false
		for _, char := range upperBytes {
			if strings.Contains(password, string(char)) {
				valid = true
				break
			}
		}

		if !valid {
			return false
		}
	}

	if os.Getenv("USE_NUMBER") == "true" {
		valid := false
		for _, char := range numBytes {
			if strings.Contains(password, string(char)) {
				valid = true
				break
			}
		}

		if !valid {
			return false
		}
	}

	return true
}

/*
Check required validation for Project ID in epic

params: fl validator.FieldLevel Field level of the validator

return: bool The result of the validation
*/
func CheckCustomRequired(fl validator.FieldLevel) bool {
	return !fl.Field().IsZero()
}

/*
Validation for Project ID in epic when updating (because cannot change epic from project to project)

params: fl validator.FieldLevel Field level of the validator

return: bool The result of the validation (always true)
*/
func DummyValidation(fl validator.FieldLevel) bool {
	return true
}

/*
Encrypt the generated token to send to the client

params: plaintext []byte The token to encrypt

key []byte The key to encrypt the token

return: []byte The encrypted token

error The error if the encryption fails
*/
func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

func SendNewUserEmail(receiver string, name, username, password string) bool {
	// Prepare emails
	var confirmationBody, usernameBody, passwordBody bytes.Buffer
	confirmationTemplate, parseErr := template.ParseFiles("./config/confirmationEmail.html")
	if parseErr != nil {
		fmt.Println(parseErr)
		return false
	}
	executeErr := confirmationTemplate.Execute(&confirmationBody, struct{ Name string }{Name: name})
	if executeErr != nil {
		fmt.Println(executeErr)
		return false
	}

	usernameTemplate, parseErr := template.ParseFiles("./config/usernameEmail.html")
	if parseErr != nil {
		fmt.Println(parseErr)
		return false
	}
	executeErr = usernameTemplate.Execute(&usernameBody, struct {
		Name     string
		Username string
	}{
		Name:     name,
		Username: username,
	})
	if executeErr != nil {
		fmt.Println(executeErr)
		return false
	}

	passwordTemplate, parseErr := template.ParseFiles("./config/passwordEmail.html")
	if parseErr != nil {
		fmt.Println(parseErr)
		return false
	}
	executeErr = passwordTemplate.Execute(&passwordBody, struct {
		Name     string
		Password string
	}{
		Name:     name,
		Password: password,
	})
	if executeErr != nil {
		fmt.Println(executeErr)
		return false
	}

	// Create emails with the templates
	confirmationMail := gomail.NewMessage()
	confirmationMail.SetHeader("From", os.Getenv("EMAIL_ADDRESS"))
	confirmationMail.SetHeader("To", receiver)
	confirmationMail.SetHeader("Subject", "Mantle Management - Confirmation Email")
	confirmationMail.SetBody("text/html", confirmationBody.String())

	usernameEmail := gomail.NewMessage()
	usernameEmail.SetHeader("From", os.Getenv("EMAIL_ADDRESS"))
	usernameEmail.SetHeader("To", receiver)
	usernameEmail.SetHeader("Subject", "Mantle Management - New Account Information Email 1")
	usernameEmail.SetBody("text/html", usernameBody.String())

	passwordEmail := gomail.NewMessage()
	passwordEmail.SetHeader("From", os.Getenv("EMAIL_ADDRESS"))
	passwordEmail.SetHeader("To", receiver)
	passwordEmail.SetHeader("Subject", "Mantle Management - New Account Information Email 2")
	passwordEmail.SetBody("text/html", passwordBody.String())

	// Send emails
	dialer := gomail.NewDialer(os.Getenv("EMAIL_HOST"), 587, os.Getenv("EMAIL_ADDRESS"), os.Getenv("EMAIL_PASSWORD"))

	dialErr := dialer.DialAndSend(confirmationMail)
	if dialErr != nil {
		fmt.Println(dialErr)
		return false
	}

	dialErr = dialer.DialAndSend(usernameEmail)
	if dialErr != nil {
		fmt.Println(dialErr)
		return false
	}

	dialErr = dialer.DialAndSend(passwordEmail)
	if dialErr != nil {
		fmt.Println(dialErr)
		return false
	}

	return true
}

func SendEmailResetPassword(receiver, name, password string) bool {
	// Prepare email
	var resetPasswordBody bytes.Buffer
	resetPasswordTemplate, parseErr := template.ParseFiles("./config/resetPasswordEmail.html")
	if parseErr != nil {
		fmt.Println(parseErr)
		return false
	}
	executeErr := resetPasswordTemplate.Execute(&resetPasswordBody, struct{ Name, Password string }{Name: name, Password: password})
	if executeErr != nil {
		fmt.Println(executeErr)
		return false
	}

	// Create email with the templates
	resetPasswordEmail := gomail.NewMessage()
	resetPasswordEmail.SetHeader("From", os.Getenv("EMAIL_ADDRESS"))
	resetPasswordEmail.SetHeader("To", receiver)
	resetPasswordEmail.SetHeader("Subject", "Mantle Management - Reset Password")
	resetPasswordEmail.SetBody("text/html", resetPasswordBody.String())

	// Send email
	dialer := gomail.NewDialer(os.Getenv("EMAIL_HOST"), 587, os.Getenv("EMAIL_ADDRESS"), os.Getenv("EMAIL_PASSWORD"))
	dialErr := dialer.DialAndSend(resetPasswordEmail)
	if dialErr != nil {
		fmt.Println(dialErr)
		return false
	}

	return true
}
