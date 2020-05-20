package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var templates = template.Must(template.ParseFiles("img.html", "upload.html"))

var validPath = regexp.MustCompile("^/(edit|save|img)/([a-zA-Z0-9]+)$")

type Page struct {
	ID   string
	Date string
}

func imgHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	ID := m[2]

	fileInfo, err := os.Stat("images/" + ID + ".png")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	date := fileInfo.ModTime().String()
	p := &Page{ID, date}
	err = templates.ExecuteTemplate(w, "img.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "upload.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	// Generate unique ID using current time in nanoseconds
	// converted to hexadecimal
	ID := fmt.Sprintf("%x", time.Now().UnixNano())

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("img")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	filepath := fmt.Sprintf("images/" + ID + ".png")

	imageFile, err := os.Create(filepath)

	if err != nil {
		fmt.Println("Error creating image file")
		return
	}

	defer imageFile.Close()

	imageFile.Write(fileBytes)

	// redirect to uploaded image
	http.Redirect(w, r, "/img/"+ID, http.StatusFound)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	return port
}

func main() {
	if _, err := os.Stat("/images/"); os.IsNotExist(err) {
		os.Mkdir("images", 0600)
	}

	http.HandleFunc("/img/", imgHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/save/", saveHandler)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	log.Fatal(http.ListenAndServe(":"+getPort(), nil))
}
