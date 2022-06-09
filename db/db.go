package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"os"
	"strconv"
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

	db.AutoMigrate(&Post{})
	db.AutoMigrate(&Image{})
	db.AutoMigrate(&ErrLogs{})

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

	result := Database.Db.Preload("Images").Find(&post, "title = ?", title)
	if result.Error != nil {
		UppendErrorWithPath(result.Error)
		logrus.Info("Could not find post", result.Error)
		return nil, result.Error
	}
	return post, nil
}

func GetImagesByPost(postId uint) ([]Image, error) {

	var images []Image

	result := Database.Db.Find(&images, "post_id = ?", postId)
	if result.Error != nil {
		UppendErrorWithPath(result.Error)
		logrus.Info("Could not find post", result.Error)
		return nil, result.Error
	}
	return images, nil

}

func CreateHTML(title string, whoTookMe string) (string, error) {

	post, err := GetPostByTitle(title)
	if err != nil {
		UppendErrorWithPath(err)
		logrus.Info("Could not find post to create HTML", err)
		return "", err
	}

	images, err := GetImagesByPost(post.ID)
	if err != nil {
		UppendErrorWithPath(err)
		logrus.Info("Could not find post to create HTML", err)
		return "", err
	}

	//  {{range .Images}}
	//               <img src={{ .Name}}>
	//           {{end}}

	//An HTML template
	var tmpl = `
	<html>
<head>
<title>{{.Title}}</title>

</head>
<body>

<img src={{  (index  .Images 0).Name }} >


{{.Text}}

{{range .Images}}
{{if index 3}}{{break}}{{end}}
{{continue}}

<img src={{ .Name}}>
{{end}}
</body>
</html>
`
	//{{if index  .Name 0}}{{break}}{{end}}
	//{{continue}}
	a := Post{
		Title:  post.Title,
		Text:   post.Text,
		Images: images,
	}

	for i, _ := range images {
		if i == 0 {
			continue
		}
		tmpl = tmpl + "<html> <img src={{  (index $.Images" + " " + strconv.Itoa(i) + ").Name }} > </html>"
	}

	// Make and parse the HTML template
	t, err := template.New(post.Title).Parse(tmpl)
	if err != nil {
		logrus.Info("Error occurred while creating new template", err)
		UppendErrorWithPath(err)
		return "", err
	}

	file, err := os.Create(post.Title + ".html")
	if err != nil {
		logrus.Info("Error occurred while creating file", err)
		UppendErrorWithPath(err)
		return "", err
	}

	err = t.ExecuteTemplate(file, post.Title, a)
	if err != nil {
		logrus.Info("Error occurred while updating file data", err)
		UppendErrorWithPath(err)
		return "", err
	}

	//err = t.ExecuteTemplate(file, post.Title, images)
	//if err != nil {
	//	logrus.Info("Error occurred while updating file data", err)
	//	UppendErrorWithPath(err)
	//	return "", err
	//}

	post.WhoTookMe = whoTookMe
	result := Database.Db.Save(&post)
	if result.Error != nil {
		logrus.Info("Error occurred while updating post", err)
		UppendErrorWithPath(result.Error)
	}

	path := "./" + post.Title + ".html"
	return path, nil
}
