package main

import(
    clean "github.com/microcosm-cc/bluemonday"
	md "github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
	"text/template"
	"io/ioutil"
	"strconv"
	"strings"
	"bytes"
	"time"
	"log"
	"os"
)

type BlogMeta struct {

	Name string
	Tagline string
	Author string

}

type PostMeta struct {

	Unixdate string
	Date time.Time
	DateString string
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
	Tag string
}

type Blog struct {

	Posts []Post
	Data BlogMeta

}

type TagSearch struct {

	Name string
	Posts []Post

}

func LocationInData(tags []TagSearch, tag string) int{

	location := -1

	for i, tagSearch := range tags {

		if tagSearch.Name == tag {
			location = i
		}

	}

	return location

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

func BuildYaml(key string, value string) string{

	yamlArray := []string{key, value}
	return strings.Join(yamlArray, ": ")

}

func timestampToSTring(timestamp int64) string{

	return strconv.Itoa(int(timestamp))

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

	log.Println(Data.Unixdate)

	if Data.Unixdate == "" {

		Data.Date = time.Now()

		f, err := os.OpenFile(strings.Join(fileLocation, ""), os.O_RDWR, 0755)

		Handle(err) 

		yamlDate := BuildYaml("unixdate", timestampToSTring(Data.Date.Unix()))

		defer f.Close()

		_, err = f.Write([]byte(strings.Join([]string{yamlDate, string(input)}, "\n")))

		Handle(err)

		f.Close()

	}else{

		intTime, err := strconv.Atoi(Data.Unixdate)

		Handle(err)

		Data.Date = time.Unix(int64(intTime), 0)

	}

	Data.DateString = Data.Date.Format("January 2, 2006")
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

func BuildTagSearch(Posts []Post) []TagSearch{

	tags := []TagSearch{}

	for i, _ := range Posts {

		for _, tag := range Posts[i].Data.Tags {

		loc := LocationInData(tags, tag)

		if loc != -1{

			tags[loc].Posts = append(tags[loc].Posts, Posts[i])

		}else{

			thisTag := TagSearch{
				Name: tag,
			}

			thisTag.Posts = append(thisTag.Posts, Posts[i])
			tags = append(tags, thisTag)

		}

		}

	}

	return tags

}

func BuildTagSearchPages(Config BlogMeta, Tags []TagSearch) {

	for _, tag := range Tags {

		TagTemplate, err := template.ParseFiles("templates/layout.html", "templates/tag.html")

		Handle(err)

		title := []string{tag.Name, Config.Name}

		tagPage := Page {
			Posts: tag.Posts,
			Data: Config,
			Title: strings.Join(title, " | "),
			Tag: tag.Name,
		}

		var out bytes.Buffer

		err = TagTemplate.ExecuteTemplate(&out, "layout.html", tagPage)

		Handle(err)

		url := []string{"output/tags/", tag.Name, ".html"}

		err = ioutil.WriteFile(strings.Join(url, ""), out.Bytes(), 0755)
		
		Handle(err)

	}

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

	Tags := BuildTagSearch(Posts)

	BuildTagSearchPages(Config, Tags)

}
 
func main(){

	Build()

}