package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

/*
[1]
- Define data structure for wiki. Here it describe how page data will be stored in memory

- type []byte means "a byte slice"

- Body element is a []byte rather than string because that is the type expected by the io libraries

*/
type Page struct {
	Title string
	Body  []byte
}

/*
[2: Save the Page's Body to a text file]
- This method's signature reads: "This is a method named save that takes as its receiver p, a pointer to Page . It takes no parameters, and returns a value of type error."

- This method will save the Page's Body to a text file. For simplicity, we will use the Title as the file name.

- The save method returns an error value because that is the return type of WriteFile

- .WriteFile is a standard library function that writes a byte slice to a file

- The save method returns the error value, to let the application handle it should anything go wrong while writing the file.

- If all goes well, Page.save() will return nil (the zero-value for pointers, interfaces, and some other types).

- The octal integer literal 0600, passed as the third parameter to WriteFile, indicates that the file should be created with read-write permissions for the current user only.

*/

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*
[3: To load pages]
- constructs the file name from the title parameter

- reads the file's contents into a new variable body

- returns a pointer to a Page literal constructed with the proper title and body values and also error

- io.ReadFile returns []byte and error.

- Callers of this function can now check the second parameter; if it is nil then it has successfully loaded a Page. If not, it will be an error that can be handled by the caller

*/
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

/*
[7]
- The function homeHandler is of the type http.HandlerFunc. It takes an http.ResponseWriter and an http.Request as its arguments.

- An http.ResponseWriter value assembles the HTTP server's response; by writing to it, we send data to the HTTP client.

- An http.Request is a data structure that represents the client HTTP request.

- r.URL.Path is the path component of the request URL.

- The trailing [1:] means "create a sub-slice of Path from the 1st character to the end." This drops the leading "/" from the path name.
*/
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "Hello world. This is the main page")
	// fmt.Fprintf(w, "Hi there, I am testing. Here's the path: %s", r.URL.Path[1:])
	p, _ := loadPage("homePage")
	renderTemplate(w, "home", p)
}

/*
[10]
*/

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

/*
[8: handle URLs prefixed with "/view/"]
- create a handler, viewHandler that will allow users to view a wiki page.

- this function extracts the page title from r.URL.Path

- The Path is re-sliced with [len("/view/"):] to drop the leading "/view/" component of the request path. This is because the path will invariably begin with "/view/", which is not part of the page's title.

- The function then loads the page data, formats the page with a string of simple HTML, and writes it to w, the http.ResponseWriter.

- Instead, if the requested Page doesn't exist, it should redirect the client to the edit Page using http.Redirect

- The http.Redirect function adds an HTTP status code of http.StatusFound (302) and a Location header to the HTTP response.
*/
func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

/*
[9]
-  template.ParseFiles will read the contents of edit.html and return a *template.Template.

- The method t.Execute executes the template, writing the generated HTML to the http.ResponseWriter.
*/
func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}
func main() {
	/*
		[4: Testing]
	*/
	p1 := &Page{Title: "Test Page", Body: []byte("This is a sample page")}
	p1.save()
	p2, _ := loadPage("Test Page")
	fmt.Println(string(p2.Body))

	/*
		[5]
		- http.HandleFunc, which tells the http package to handle all requests to the web root ("/") with handler.
	*/
	http.HandleFunc("/", homeHandler)
	/*
		[7: Add in request handler for viewHandler]
	*/
	http.HandleFunc("/view/", viewHandler)
	/*
		[8]
	*/
	http.HandleFunc("/edit/", editHandler)
	// http.HandleFunc("/save/", saveHandler)
	/*
		[6]
		- It then calls http.ListenAndServe, specifying that it should listen on port 8080 on any interface (":8080"). (Don't worry about its second parameter, nil, for now.) This function will block until the program is terminated.

		- ListenAndServe always returns an error, since it only returns when an unexpected error occurs. In order to log that error we wrap the function call with log.Fatal.
	*/
	log.Fatal(http.ListenAndServe(":8080", nil))
}
