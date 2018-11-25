package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
)

var (
	slackTokens = []string{
		"70XqnEL12zOlA08Fo0lraciE",
		"ayYWtEzhfqh5GcXdEqrD3H3h",
		"7b5WbqiybRqPDRTm2e9GvTUL",
		"x56o3ZQzti2l7YEb7ntRu4gE",
	}
	definitionLimit = 5
)

func isAuthorizedToken(token string) bool {
	for _, tok := range slackTokens {
		if tok == token {
			return true
		}
	}
	return false
}

type defineRequest struct {
	Text    string
	Token   string
	Trigger string
}

type Term string

func (t Term) Raw() string {
	return strings.Split(string(t), ": ")[1]
}

func (t Term) String() string {
	term := t.Raw()
	st := strings.Split(strings.ToLower(term), " ")
	return strings.Join(st, "-")
}

type slackResponse struct {
	Text string `json:"text"`
}

func defineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		fmt.Fprint(w, "only POST requests supported")
		return
	}
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "something is bad with your data: %s", err)
		return
	}
	data := defineRequest{
		Token:   r.FormValue("token"),
		Text:    r.FormValue("text"),
		Trigger: r.FormValue("trigger_word"),
	}
	// Make request to dictionary site
	if !isAuthorizedToken(data.Token) {
		w.WriteHeader(403)
		fmt.Fprint(w, "not authorized")
		return
	}
	t := Term(data.Text)
	log.Printf("Going to search for %s\n", t)
	dictURL := fmt.Sprintf("https://en.oxforddictionaries.com/definition/%s", t)
	resp, err := soup.Get(dictURL)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return
	}
	doc := soup.HTMLParse(resp)
	main := doc.Find("section", "class", "gramb")
	defs := main.FindAll("span", "class", "ind")
	grammar := main.Find("span", "class", "pos")

	var o string
	for i, d := range defs {
		if i < definitionLimit {
			o += fmt.Sprintf("%d. %s\n", i+1, d.Text())
		}
	}
	if len(o) > 0 {
		o = fmt.Sprintf("Definitions for *%s* - _%s_\n\n", t.Raw(), grammar.Text()) + o
		o += fmt.Sprintf("\n_Brought to you by <%s|English Oxford Dictionaries>_", dictURL)
	}
	if len(defs) == 0 {
		o = fmt.Sprintf("Couldn't find anything for %s!", data.Text)
	}

	json.NewEncoder(w).Encode(slackResponse{Text: o})
}

func main() {
	http.HandleFunc("/", defineHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	log.Printf("Error? %s", http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}