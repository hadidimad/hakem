package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var templates *template.Template

func InitRender() {
	templates = template.New("")
	funcMap := template.FuncMap{
		"getUserImage": func(id int) string {
			onuser := getUserByID(id)
			image, err := ioutil.ReadFile("./userImages/" + onuser.Username)
			if err != nil {
				fmt.Println(err)
				image, _ := ioutil.ReadFile("./userImages/noimage")
				ioutil.WriteFile("./statics/userImages/"+onuser.Username, image, 0644)
				return "/static/userImages/" + onuser.Username
			}
			ioutil.WriteFile("./statics/userImages/"+onuser.Username, image, 0644)
			return "/static/userImages/" + onuser.Username
		},
		"getCardImage": func(Card card) string {
			return getCardImg(Card)
		},
	}
	templates.Funcs(funcMap)
	err := filepath.Walk("./view/", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templates.ParseFiles(path)
		}
		return err
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func Render(w io.Writer, name string, m interface{}) {
	err := templates.ExecuteTemplate(w, name, m)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
