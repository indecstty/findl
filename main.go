package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	var year, filename, email, password string
	flag.StringVar(&year, "year", "", "Vuosiluvun kaksi viimeistä numeroa, esim. 19")
	flag.StringVar(&filename, "file", "db.csv", "Tietokannan csv-export")
	flag.Parse()

	baseURL := "https://indecs.fi/findecs" + year
	urlString := baseURL + "/users/login"

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Sähköposti: ")
	email, _ = reader.ReadString('\n')
	email = strings.TrimSpace(email)
	fmt.Print("Salasana: ")
	pwBytes, _ := terminal.ReadPassword(0)
	password = string(pwBytes)
	fmt.Println("\nAuthenticating against", urlString)

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 && via[0].URL.String() == urlString {
				return nil
			}
			return errors.New("authentication failed")
		},
		Jar: jar,
	}

	res, err := client.PostForm(urlString, url.Values{
		"data[User][email]":    []string{email},
		"data[User][password]": []string{password},
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Status)

	fmt.Printf("Reading DB file %s...\n", filename)
	fd, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	cr := csv.NewReader(fd)
	claims, err := cr.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, claim := range claims {
		id := claim[0]
		receipts := strings.Split(claim[1], ";")

		fmt.Printf("Claim %s ", id)

		f := startHTML(claim[2])

		res, err := client.Get(fmt.Sprintf(baseURL+"/CostClaims/view/%s/print", id))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res.Status)
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		body := "<table>" + strings.Split(string(bodyBytes), "table>")[1] + "table>"
		f.WriteString(body)

		for _, receipt := range receipts {
			if receipt == "" {
				continue
			}
			fmt.Printf("> Receipt %s\n", receipt)
			res, err = client.Get(baseURL + "/files/receipts/" + receipt)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("GET %s: %s\n", res.Request.URL, res.Status)

			img, err := os.Create("output/" + receipt)
			if err != nil {
				log.Fatal(err)
			}
			io.Copy(img, res.Body)
			img.Close()

			receipt = url.PathEscape(receipt)

			f.WriteString(fmt.Sprintf(`<img style="max-height: 1000px; max-width: 200mm; margin-bottom: 20px;" src="%s" />`, receipt))
		}

		f.WriteString("</body></html>")
		f.Close()
	}
}

func startHTML(filename string) *os.File {
	f, err := os.Create(fmt.Sprintf("output/%s.html", filename))
	if err != nil {
		log.Fatal(err)
	}

	f.WriteString(`
		<html>
		<head>
			<meta http-equiv='content-type' content='text/html; charset=utf-8'>
			<style>
				body {
					max-width: 200mm;
				}
				td{
					border:1px solid black;
					border-left:none;
					border-right:none;
					padding:0;
				}
				tr{
					padding:0;
				}
				table{
					padding:0;
					border-collapse: collapse;
					width:100%;
					font-size:1.1em;
				}
				img{
					max-height:100px;
					padding:10px;
				}
				.legend{
					border-bottom:none;
					font-size:0.8em;
				}
				.content{
					border-top:none;
					padding-left:0.7em;
				}
				.signature{
					height:10mm;
					min-width:75mm;
				}
				.note-area{
					border-bottom:none;
				}
				.total-cell{
					border-top:2px solid black;
					font-weight:bold;
					line-height:2em;
				}
				.divider td{
					line-height:2em;
					font-size:0.8em;
					border-bottom:0;
				}
				.gutter{
					width:10mm;
					border:none;
				}
				.first-row td{
					border-top:0;
				}
				.first-row img {
					display: none;
				}
				.no-border{
					border:0;
				}
			</style>
		</head>
		<body>	
	`)

	return f
}
