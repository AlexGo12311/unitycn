package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func home_page(w http.ResponseWriter, r *http.Request) {
	mao := User{Name: "Mao", Age: 25, Avg_raiting: 2.5, Happiness: 3.0}
	tmpl, _ := template.ParseFiles("templates/home_page.html")
	tmpl.Execute(w, mao)
}

func news_page(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to news page!")
}

func handleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/news/", news_page)
	http.ListenAndServe(":8080", nil)
}

type User struct {
	Name                   string
	Age                    uint
	Avg_raiting, Happiness float64
}

func (u User) getAllInfo() string {
	return fmt.Sprintf("User name is: %s. He is %d years old.", u.Name, u.Age)
}

func main() {
	handleRequest()
}
