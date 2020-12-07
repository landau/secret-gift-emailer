package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"mime/quotedprintable"
	"net/smtp"
	"os"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

const body = `
<h1>Happy Hanukkah!</h1>
<br />
<h2>Here is your Secret Hanukkah Exchannge receipient:</h2>
<br />
Name:  {{.Name}}
<br />
Email: {{.Email}}
<br />
Address: {{.Address}}
<br />
Please send a gift in the range of $20 by Dec 17th.
`

func createEmailHeaders(from, to, subject string) string {
	headers := make(map[string]string)

	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject

	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", "text/html")
	headers["Content-Transfer-Encoding"] = "quoted-printable"
	headers["Content-Disposition"] = "inline"

	headerMessage := ""
	for key, value := range headers {
		headerMessage += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	return headerMessage
}

func parseTemplate(t *template.Template, data interface{}) (string, error) {
	var output bytes.Buffer
	err := t.Execute(&output, data)

	if err != nil {
		log.Println("executing template:", err)
		return "", err
	}

	return output.String(), nil
}

func createEmailBody(body string) string {
	var buf bytes.Buffer
	quotedWriter := quotedprintable.NewWriter(&buf)
	quotedWriter.Write([]byte(body))
	quotedWriter.Close()
	return buf.String()
}

type emailConfig struct {
	ToEmail   string
	FromEmail string
	Subject   string
	Body      string
	Person    person // The person to be templated in the body of the email
	Password  string
}

func sendGMail(conf emailConfig) error {
	host := "smtp.gmail.com"
	auth := smtp.PlainAuth("", conf.FromEmail, conf.Password, host)

	addr := host + ":587"
	to := []string{conf.ToEmail}

	headers := createEmailHeaders(conf.FromEmail, conf.ToEmail, conf.Subject)
	body := createEmailBody(conf.Body)
	msg := []byte(headers + "\r\n" + body)

	err := smtp.SendMail(addr, auth, conf.FromEmail, to, msg)
	return err
}

// ReadCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
func readCsv(filename string) ([][]string, error) {
	// Open CSV file
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		return [][]string{}, err
	}

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()

	if err != nil {
		return nil, err
	}

	// Scrap the keys
	return lines[1:], nil
}

type person struct {
	Name    string
	Email   string
	Address string
}

func readPersonCsv(filename string) ([]person, error) {
	lines, err := readCsv(filename)

	if err != nil {
		return nil, err
	}

	people := make([]person, 0)

	for _, line := range lines {
		people = append(people, person{
			Name:    line[0],
			Email:   line[1],
			Address: line[2],
		})
	}

	return people, nil
}

func assignReceipients(people []person) map[person]person {
	output := make(map[person]person)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(people), func(i, j int) { people[i], people[j] = people[j], people[i] })

	for i, person := range people {
		output[person] = people[(i+1)%len(people)]
	}

	return output
}

func main() {
	email := os.Getenv("GMAIL_EMAIL")
	subject := "Secret Hanukkah Gift Exchange"

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		log.Fatalf("Failed to read in passowrd: %s", err)
	}

	password := string(bytePassword)

	fileName := os.Getenv("CSV")
	people, err := readPersonCsv(fileName)

	if err != nil {
		log.Fatalf("Failed to read in csv: %s", err)
	}

	emailTemplate := template.Must(template.New("email").Parse(body))

	for fromPerson, toPerson := range assignReceipients(people) {
		emailBody, err := parseTemplate(emailTemplate, toPerson)

		if err != nil {
			log.Fatalf("Error creating email body: %s", err)
		}

		conf := emailConfig{
			FromEmail: email,
			Password:  password,
			ToEmail:   fromPerson.Email,
			Subject:   subject,
			Body:      emailBody,
		}

		// Re-enable log if you'e interested in seeing the secret match
		if os.Getenv("DEBUG") == "true" {
			log.Printf("%+v", conf)
		} else {
			log.Println("NOT HEREE!!!!")
			// err = sendGMail(conf)
			if err != nil {
				log.Fatalf("Error sending email: %s", err)
			}

			log.Print("Email Sent Successfully")
		}

	}
}
