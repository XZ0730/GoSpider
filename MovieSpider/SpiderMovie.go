package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type M_Comment struct {
	Id           int64  `gorm:"primary_key;auto_increment"`
	MajorComment string `gorm:"type:text"`
}

var M_comments []M_Comment

func httpget(url string) (result string, err error) {

	resp, err2 := http.Get(url)
	if err2 != nil {
		err = err2
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024*200)
	for {
		n, _ := resp.Body.Read(buf)
		if n == 0 {
			break
		}

		result += string(buf)
	}
	return
}
func GetSpecialData(htmlContent string, selector string) (string, error) { //css选择器
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	var str string
	dom.Find(selector).Each(func(i int, selection *goquery.Selection) {
		str = selection.Text()
	})
	return str, nil
}

func Spiderhtml(url string) {
	var Comment_temp M_Comment
	//参数信息 全局：[]M_Comment

	result2, _ := httpget("https://api.bilibili.com/pgc/season/episode/web/info?ep_id=199612")
	re := regexp.MustCompile(`"reply":(?s:(.*?)),"`) //总评论数
	str := re.FindAllStringSubmatch(result2, -1)
	fmt.Println(str[0][1])

	ComNum, err3 := strconv.Atoi(str[0][1])
	fmt.Println(ComNum)
	if err3 != nil {
		panic(err3)
	}
	fmt.Println(ComNum)
	var k int64 = 1 //评论索引
	for i := 1; i <= ComNum/20; i++ {
		//str1 := "https://api.bilibili.com/x/v2/reply/main?callback=jQueryjsonp=jsonp&next=i&type=1&oid=21071819&mode=3&plat=1&_="
		str1 := "https://api.bilibili.com/x/v2/reply/main?callback=jQueryjsonp=jsonp&next=" + strconv.Itoa(i) + "&type=1&oid=21071819&mode=3&plat=1&_="
		//https://api.bilibili.com/x/v2/reply/main?callback=jQueryjsonp=jsonp&next=3&type=1&oid=21071819&mode=3&plat=1&_=1669618037648
		result3, err2 := httpget(str1) //
		if err2 != nil {
			panic(err2)
		}

		re1 := regexp.MustCompile(`"content":{"message":"(?s:(.*?))"`)
		s := re1.FindAllStringSubmatch(result3, -1)

		for _, v := range s {
			Comment_temp.MajorComment = v[1]
			M_comments = append(M_comments, Comment_temp)
			// fmt.Println(v[1])
			k++
		}

	}

}
func DoWoke() {

	url := "https://www.bilibili.com/bangumi/play/ss12548"
	Spiderhtml(url)

}

func main() {
	DoWoke()
	db1, err := gorm.Open("mysql", "root:111111@(127.0.0.1:3306)/txxt?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	// 创建表 自动迁移 (把结构体和数据表进行对应)
	db1.AutoMigrate(&M_Comment{})
	for _, v := range M_comments {
		db1.Create(&v)
	}

}
