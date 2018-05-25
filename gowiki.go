package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
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
[12: Template caching]
- create a global variable named templates, and initialize it with ParseFiles

- The function template.Must is a convenience wrapper that panics when passed a non-nil error value, and otherwise returns the *Template unaltered.

- A panic is appropriate here; if the templates can't be loaded the only sensible thing to do is exit the program.

- ParseFiles function takes any number of string arguments that identify our template files, and parses those files into templates that are named after the base file name.
*/
var templates = template.Must(template.ParseFiles("home.html", "edit.html", "view.html"))

/*
[10]
- Error Handling:
A better solution is to handle the errors and return an error message to the user. That way if something does go wrong, the server will function exactly how we want and the user can be notifie

- The http.Error function sends a specified HTTP response code (in this case "Internal Server Error") and error message.
*/

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

/*
[8: handle URLs prefixed with "/view/"]
- create a handler, viewHandler that will allow users to view a wiki page.

- this function extracts the page title from r.URL.Path

- The Path is re-sliced with [len("/view/"):] to drop the leading "/view/" component of the request path. This is because the path will invariably begin with "/view/", which is not part of the page's title.

- The function then loads the page data, formats the page with a string of simple HTML, and writes it to w, the http.ResponseWriter.

- Instead, if the requested Page doesn't exist, it should redirect the client to the edit Page using http.Redirect

- The http.Redirect function adds an HTTP status code of http.StatusFound (302) and a Location header to the HTTP response.
*/
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
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
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

/*
[11: The function saveHandler will handle the submission of forms located on the edit pages.]

- The page title (provided in the URL) and the form's only field, Body, are stored in a new Page.

- The save() method is then called to write the data to a file, and the client is redirected to the /view/ page.

- The value returned by FormValue is of type string. We must convert that value to []byte before it will fit into the Page struct. We use []byte(body) to perform the conversion.


*/
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

/*
[14: function literals and closures]
- The closure returned by makeHandler is a function that takes an http.ResponseWriter and http.Request (in other words, an http.HandlerFunc). AND returns a function of type http.HandlerFunc

- The closure extracts the title from the request path, and validates it with the TitleValidator regexp.

-If the title is invalid, an error will be written to the ResponseWriter using the http.NotFound function.

-If the title is valid, the enclosed handler function fn will be called with the ResponseWriter, Request, and title as arguments.
*/

/*
[13: Validation for Regexp]
- If the title is valid, it will be returned along with a nil error value. If the title is invalid, the function will write a "404 Not Found" error to the HTTP connection, and return an error to the handler. To create a new error, we have to import the errors package.
*/

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}
func main() {
	/*
		[5]
		- http.HandleFunc, which tells the http package to handle all requests to the web root ("/") with handler.
	*/
	http.HandleFunc("/", homeHandler)
	/*
		[7: Add in request handler for viewHandler]
	*/
	http.HandleFunc("/view/", makeHandler(viewHandler))
	/*
		[8]
	*/
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	/*
		[6]
		- It then calls http.ListenAndServe, specifying that it should listen on port 8080 on any interface (":8080"). (Don't worry about its second parameter, nil, for now.) This function will block until the program is terminated.

		- ListenAndServe always returns an error, since it only returns when an unexpected error occurs. In order to log that error we wrap the function call with log.Fatal.
	*/
	log.Fatal(http.ListenAndServe(":8080", nil))
}
