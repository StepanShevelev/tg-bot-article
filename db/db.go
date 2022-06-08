package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"log"
	"os"
)

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

func ConnectToDb() {
	dsn := "host=localhost port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=Europe/Moscow"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		UppendErrorWithPath(err)
		logrus.Fatal("Failed to connect to the database! \n", err)
	}

	Database = DbInstance{
		Db: db,
	}
}

func GetPosts() (map[int]string, error) {
	var posts []*Post

	//var postArr []string
	postMap := make(map[int]string)
	result := Database.Db.Find(&posts, "who_took_me = ?", "")
	if result.Error != nil {
		UppendErrorWithPath(result.Error)
		logrus.Info("Could not find post", result.Error)
		return nil, result.Error
	}

	for i, post := range posts {
		postMap[i] = post.Title
	}
	return postMap, nil
}

func GetPostByTitle(title string) (*Post, error) {

	var post *Post

	result := Database.Db.Find(&post, "title = ?", title)
	if result.Error != nil {
		UppendErrorWithPath(result.Error)
		logrus.Info("Could not find post", result.Error)
		return nil, result.Error
	}
	return post, nil
}

//*tgbotapi.Document
//title string

func CreateHTML(title string) *os.File {
	//var docDoc *tgbotapi.Document

	post, err := GetPostByTitle(title)
	if err != nil {
		UppendErrorWithPath(err)
		logrus.Info("Could not find post to create HTML", err)
	}

	//An HTML template
	const tmpl = `

<html>
<head>
<title>{{.Title}}</title>
</head>
<body>
{{.Text}}
</body>
</html>
`

	// Make and parse the HTML template
	t, err := template.New(post.Title).Parse(tmpl)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(post.Title + ".html")
	if err != nil {
		log.Panic(err)
	}
	t.ExecuteTemplate(file, post.Title, post)

	//defer os.Remove(file.Name())
	// Render the data and output using standard output
	//t.Execute(os.Stdout, post)

	return file
}
