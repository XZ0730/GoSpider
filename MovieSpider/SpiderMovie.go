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
			fmt.Println(k)
			Comment_temp.MajorComment = v[1]
			M_comments = append(M_comments, Comment_temp)
			fmt.Println(v[1])
			k++
		}
		//因为b站的主评论每二十条刷新，主评论又是再js中
		//js的url滚动刷新后才能显示，所以这里就只爬几十个，下面这个i可以控制爬更多，不过得去b站评论区滚动刷新后才能爬出来，不滚动刷新去爬更多会报错
		//我技术有限，本来是想和爬fzu网站一样并发的，但是这些数据全是都是js渲染的动态数据，以我掌握的知识感觉做起来太麻烦了，
		//所以就只能手动抓包，直接把主评论和子评论混在一起存数据库，还有就是这个js中的主评论下的子评论只能爬到一两个，其他的爬不到不知道为啥，
		//我看了js文件里的内容，也确实没有更多的子评论，手动点了翻页刷新之后也只显示这些，，所以每个主评论的子评论都只爬了一点
		if i == 1 { //i一般是切 1，因为初始化是二十条主评论，滚动刷新一次 可以加一，方便测试所以就先为一，差不多滚动刷新一次是多七十或八十多条评论
			break
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
