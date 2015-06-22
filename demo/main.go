package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const addr = "https://api.meetup.com/2/open_events"

var key = os.Getenv("MEETUP_KEY")

func getDescriptions(body []byte) ([]string, error) {
	des := []string{}
	data := make(map[string]interface{})
	if err := json.Unmarshal(body, &data); err != nil {
		return des, err
	}
	results, prs := data["results"]
	if !prs {
		return des, errors.New("No results")
	}
	for _, result := range results.([]interface{}) {
		r := result.(map[string]interface{})
		description, prs := r["description"]
		if !prs {
			continue
		}
		d := description.(string)
		des = append(des, d)
	}
	return des, nil
}

func main() {

	// CLI
	l := flag.Int("l", 20, "Limit results")
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatalln("You have to provide a text")
	}
	text := flag.Args()[0]

	// Request
	query := url.Values{}
	query.Set("key", key)
	query.Set("text", text)
	query.Set("text_format", "plain")
	query.Set("status", "past")
	query.Set("page", strconv.Itoa(*l))
	resp, err := client.Get(fmt.Sprintf("%s?%s", addr, query.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	// Response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	des, err := getDescriptions(body)
	if err != nil {
		log.Fatalln(des)
	}

	// Display
	desL := des
	if *l != 0 {
		desL = des[0:*l]
	}
	for _, d := range desL {
		fmt.Println(strings.Replace(d, "\n", "", -1))
	}
}
