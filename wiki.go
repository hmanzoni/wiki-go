package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
    "regexp"
)

type Page struct {
	Title string
	Body  []byte
	Files  []string
}

// Task: Store templates in tmpl/ and page data in data/.

var dataFolder string = "data"

func checkExistsFolder(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

func createFolder(nameFolder string) {
	check, errMsg := checkExistsFolder(nameFolder)
	if errMsg != nil {
		log.Fatal(errMsg)
	}
    if check == false {
		err := os.Mkdir(nameFolder, 0755)
		if err != nil {
			log.Fatal(err)
		}
    }
}

func (p *Page) save() error {
	filename := dataFolder + "/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

// Task: Implement inter-page linking by converting instances of [PageName] to
// <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
func readavailableFiles(nameFolder string) ([]string) {
	var myArr []string
    entries, err := os.ReadDir("./"+nameFolder)
    if err != nil {
        log.Fatal(err)
    }
    sampleRegexp := regexp.MustCompile(`.txt`)
    for _, e := range entries {
		fileName := sampleRegexp.ReplaceAllString(e.Name(), "")
		myArr = append(myArr, fileName)
    }
	return myArr
}

func loadPage(title string) (*Page, error) {
	filename := dataFolder + "/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
    fileArr := readavailableFiles(dataFolder)
	return &Page{Title: title, Body: body, Files: fileArr}, nil
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Here we will extract the page title from the Request,
        // and call the provided handler 'fn'
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// Task: Add a handler to make the web root redirect to /view/FrontPage.
func rootHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func main() {
    createFolder(dataFolder)
    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))

    log.Fatal(http.ListenAndServe(":8080", nil))
}
