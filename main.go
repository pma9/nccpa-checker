package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type certInfo struct {
	Name                            string `json:"Name"`
	CertificationStatus             string `json:"CertificationStatus"`
	CertificationMessage            string `json:"CertificationMessage"`
	CityState                       string `json:"CityState"`
	AllowCredentialRequest          bool   `json:"AllowCredentialRequest"`
	IsPANCEApplicant                bool   `json:"IsPANCEApplicant"`
	HasReportableDisciplinaryAction bool   `json:"HasReportableDisciplinaryAction"`
	CertificationMessageWeb         string `json:"CertificationMessageWeb"`
	PaID                            int    `json:"PaId"`
	CertificationProduct            int    `json:"CertificationProduct"`
	CertificationProductName        string `json:"CertificationProductName"`
	PaStatus                        int    `json:"PaStatus"`
	PaStatusName                    string `json:"PaStatusName"`
	CAQStatus                       int    `json:"CAQStatus"`
	CAQStatusName                   string `json:"CAQStatusName"`
	GraduationDate                  string `json:"GraduationDate"`
	ExpectedGraduationDate          string `json:"ExpectedGraduationDate"`
	IsSurgery                       bool   `json:"IsSurgery"`
	IsSpecialty                     bool   `json:"IsSpecialty"`
	IsCurrent                       bool   `json:"IsCurrent"`
}

type byAttrParams struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	CountryCode string `json:"countryCode"`
	StateCode   string `json:"stateCode"`
	Token       string `json:"token"`
}

type byIDParams struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

var nccpaID string
var firstName string
var lastName string
var stateCode string
var countryCode string
var tokenFile string

func init() {
	flag.StringVar(&nccpaID, "id", "", "NCCPA ID")
	flag.StringVar(&firstName, "fn", "", "First Name")
	flag.StringVar(&lastName, "ln", "", "Last Name")
	flag.StringVar(&stateCode, "sc", "CA", "State Code (Default: CA)")
	flag.StringVar(&countryCode, "cc", "USA", "Country Code (Default: USA)")
	flag.StringVar(&tokenFile, "tf", "", "Token file (Default: token.txt)")
}

func main() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	if nccpaID != "" {
		params := byIDParams{
			ID:    nccpaID,
			Token: getToken(tokenFile),
		}
		log.Printf("Starting NCCPA Checker for ID: %v\n", nccpaID)
		start(func() {
			c := getCertInfoById(params)
			checkCertStatus(c)
		})
		return
	}

	if (firstName == "") || (lastName == "") {
		log.Fatalln("If you don't pass ID, you must pass both First Name and Last Name")
	}

	params := byAttrParams{
		FirstName:   firstName,
		LastName:    lastName,
		StateCode:   stateCode,
		CountryCode: countryCode,
		Token:       getToken(tokenFile),
	}

	log.Printf("Starting NCCPA Checker for %v %v\n", firstName, lastName)
	start(func() {
		c := getCertInfoByAttr(params)
		checkCertStatus(c)
	})
}

func checkCertStatus(c certInfo) {
	if c.CertificationStatus == "Certified" {
		log.Println("YESSSS!!!! Certified! : " + c.CertificationMessage)
		sendMessage(c.CertificationMessage)
		os.Exit(0)
	} else {
		log.Println("INFO - Not yet : " + c.CertificationMessage)
	}
}

func start(certFn func()) {
	// Check once the program starts to
	// make sure we can find the person
	certFn()

	locat, error := time.LoadLocation("America/New_York")
	if error != nil {
		log.Fatalln(error)
	}

	s := gocron.NewScheduler(locat)
	s.Every(1).Day()
	s.At("06:55,07:00,07:05,07:10")
	s.At("07:55,08:00,08:05,08:10")
	s.At("08:55,09:00,09:05,09:10")
	s.Do(func() {
		certFn()
	})

	s.StartBlocking()
}

func postRequest(url string, data any) *http.Response {
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatalln(err)
	}

	return resp
}

func getCertInfoById(p byIDParams) certInfo {
	resp := postRequest("https://portal.nccpa.net/verifypac/SearchById", p)

	var res certInfo
	json.NewDecoder(resp.Body).Decode(&res)
	if res.Name == "" {
		log.Fatalf("ERROR - Nobody found with the following ID : %s\n", p.ID)
	}
	return res
}

func getCertInfoByAttr(p byAttrParams) certInfo {
	resp := postRequest("https://portal.nccpa.net/verifypac/SearchByAttributes", p)

	var res []certInfo
	json.NewDecoder(resp.Body).Decode(&res)

	if len(res) < 1 {
		log.Fatalf("ERROR - Nobody found with the following paramters : %+v\n", p)
	}

	if len(res) > 1 {
		log.Fatalf("WARN - There were two certifications found! : %+v\n", res)
	}

	return res[0]
}

func getToken(tokenFile string) string {
	tf := "token.txt"
	if tokenFile != "" {
		tf = tokenFile
	}

	token, err := os.ReadFile(tf)
	if err != nil {
		log.Fatalln(err)
	}

	return string(token)
}

func sendMessage(msg string) {
	client := twilio.NewRestClient()

	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")
	toNumber := os.Getenv("TWILIO_TO_NUMBER")

	params := &openapi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetFrom(fromNumber)
	params.SetBody(msg)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		log.Println("ERROR - Unable to send message : " + err.Error())
		return
	}

	response, _ := json.Marshal(*resp)
	log.Println("INFO - Message sent : " + string(response))
}
