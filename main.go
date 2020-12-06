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
	"time"
)

type person struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func getTemplate(p person) (string, error) {
	body := `
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

	t := template.Must(template.New("email").Parse(body))

	var output bytes.Buffer
	err := t.Execute(&output, p)

	if err != nil {
		log.Println("executing template:", err)
		return "", err
	}

	return output.String(), nil
}

func sendMail(fromPerson, toPerson person, sender, password string) {
	host := "smtp.gmail.com:587"
	auth := smtp.PlainAuth("", sender, password, "smtp.gmail.com")

	header := make(map[string]string)
	header["From"] = sender
	header["To"] = fromPerson.Email
	header["Subject"] = "Secret Hanukkah Gift Exchange"

	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", "text/html")
	header["Content-Transfer-Encoding"] = "quoted-printable"
	header["Content-Disposition"] = "inline"

	headerMessage := ""
	for key, value := range header {
		headerMessage += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	body, err := getTemplate(toPerson)
	if err != nil {
		log.Printf("Error parsing template: %s", err)
	}

	var bodyMessage bytes.Buffer
	temp := quotedprintable.NewWriter(&bodyMessage)
	temp.Write([]byte(body))
	temp.Close()

	finalMessage := headerMessage + "\r\n" + bodyMessage.String()
	status := smtp.SendMail(host, auth, sender, []string{fromPerson.Email}, []byte(finalMessage))

	if status != nil {
		log.Printf("Error from SMTP Server: %s", status)
	}

	log.Print("Email Sent Successfully")
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

	return lines[1:], nil
}

func readPersonCsv(filename string) ([]person, error) {
	lines, err := readCsv(filename)

	if err != nil {
		return nil, err
	}

	people := make([]person, 0)

	// Loop through lines & turn into object
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
	people, err := readPersonCsv("test.csv") // TODO: get from CLI

	if err != nil {
		panic(err)
	}

	email := ""    // TODO: get from CLI
	password := "" // TODO: Get from CLI securely

	for fromPerson, toPerson := range assignReceipients(people) {
		log.Printf("%v: %v\n", fromPerson.Name, toPerson.Name)
		sendMail(fromPerson, toPerson, email, password)
	}
}
