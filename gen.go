package main

import(
    clean "github.com/microcosm-cc/bluemonday"
	md "github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
	"text/template"
	"io/ioutil"
	"strings"
	"bytes"
	"time"
	"log"
)

type BlogMeta struct {

	Name string
	Tagline string
	Author string

}

type PostMeta struct {

	Date time.Time
	Title string
	Tags []string
	URL string

}

type Post struct {

	Name string
	Source string
	Output string
	Data PostMeta 

}

type Page struct {
	Posts []Post
	Post Post
	Data BlogMeta
	Title string
}

type Blog struct {

	Posts []Post
	Data BlogMeta

}

func Handle(err error) {

	if err != nil{
		log.Fatal(err)
	}

}

func BuildURL(name string) string{

	s := []string{strings.TrimSuffix(name, ".md"), ".html"}
	return strings.Join(s, "")

}

func BuildPost(filename string) Post{

	fileLocation := []string{"posts/", filename}
	input, err := ioutil.ReadFile(strings.Join(fileLocation, ""))

	Handle(err)

	var Data PostMeta

	splitFormats := strings.SplitN(string(input), "---", 2)

	configData := splitFormats[0]
	postContent := splitFormats[1]

	err = yaml.Unmarshal([]byte(configData), &Data)

	Handle(err)

	Data.Date = time.Now()
	Data.URL = BuildURL(filename)

	unsafe := md.MarkdownCommon([]byte(postContent))
	html := clean.UGCPolicy().SanitizeBytes(unsafe)

	return Post {
		Name: filename,
		Source: string(input),
		Output: string(html),
		Data: Data,
	}

}

func BuildConfig() BlogMeta {

	input, err := ioutil.ReadFile("config.yml")

	Handle(err)

	var Config BlogMeta

	err = yaml.Unmarshal(input, &Config)

	Handle(err)
	
	return Config

}

func BuildPosts(Config BlogMeta) []Post{

	files, err := ioutil.ReadDir("posts")

	Handle(err)

	postTemplate, err := template.ParseFiles("templates/layout.html", "templates/post.html")

	Handle(err)

	var posts []Post

	for i, file := range files {
		posts = append(posts, BuildPost(file.Name()))

		title := []string{posts[i].Data.Title, Config.Name}

		postPage := Page {
			Post: posts[i],
			Data: Config,
			Title: strings.Join(title, " | "),
		}
		
		var out bytes.Buffer
		err := postTemplate.ExecuteTemplate(&out, "layout.html", postPage)
		Handle(err)
		
		outputLocation := []string{"output/posts/", BuildURL(file.Name())}

		err = ioutil.WriteFile(strings.Join(outputLocation, ""), out.Bytes(), 0755)
		Handle(err)
	}

	return posts

}

func BuildIndex(Config BlogMeta, Posts []Post){

	IndexTemplate, err := template.ParseFiles("templates/layout.html", "templates/index.html")

	Handle(err)

	indexPage := Page {
		Posts: Posts,
		Data: Config,
		Title: Config.Name,
	}

	var out bytes.Buffer
	err = IndexTemplate.ExecuteTemplate(&out, "layout.html", indexPage)
	Handle(err)

	err = ioutil.WriteFile("output/index.html", out.Bytes(), 0755)
	Handle(err)

}

func Build(){

	Config := BuildConfig()

	Posts := BuildPosts(Config)

	BuildIndex(Config, Posts)

}
 
func main(){

	Build()

}