package main

import (
	"fmt"
	"io/ioutil"
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

func main() {
	p1 := &Page{Title: "Test Page", Body: []byte("This is a sample page")}
	p1.save()
	p2, _ := loadPage("Test Page")
	fmt.Println(string(p2.Body))
}
