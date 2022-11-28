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

type Arti struct {
	Id         int `gorm:"primary_key;auto_increment"`
	Author     string
	Title      string
	Content    string `gorm:"type:longText"` //longtext类型也可以text类型
	CreateTime string
	ReadNum    string `gorm:"type:varchar(20)"`
}

func (Arti) TableName() string {
	return "article"
}

var articles []Arti //存储文章

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
func httpget(url string) (result string, err error) {
	resp, err2 := http.Get(url)
	if err2 != nil {
		err = err2
		fmt.Println(err2)
		return
	} //
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

func spider2(url string) (arr *Arti, err error) {

	result, err2 := httpget(url)
	if err2 != nil {
		err = err2
		return
	}
	re1 := regexp.MustCompile(`<h3>(?s:(.*?))</h3>`)

	re2 := regexp.MustCompile(`<span>发布日期:  (?s:(.*?))</span>`)

	re3 := regexp.MustCompile(`<span>作者：(?s:(.*?))</span>`)

	if re1 == nil || re2 == nil {
		fmt.Println("[Error] regexpcmopile err")
		return
	} //<span id="dynclicks_wbnews_26221_233" name="dynclicks_wbnews_26221_233">414</span>

	var ari = new(Arti) //单例

	//正则匹配标题	发布时间 作者
	joyUrls1 := re1.FindAllStringSubmatch(result, -1)
	joyUrls2 := re2.FindAllStringSubmatch(result, -1)
	joyUrls3 := re3.FindAllStringSubmatch(result, -1)

	var title string = joyUrls1[10][1]
	var Posttime string = joyUrls2[1][1]
	var Author string = joyUrls3[1][1]
	ari.Title = title
	ari.CreateTime = Posttime
	ari.Author = Author

	//初始化选择器
	selector := "body > section > section.n_container > div > div.n_right.fr > section > form > div > div.nav01 > h6 > span:nth-child(3) > script"

	str, err3 := GetSpecialData(result, selector)
	if err3 != nil {
		err = err3
	}
	str = str[37:42] //clickid
	url2 := "https://news.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=" + str + "&owner=1779559075&clicktype=wbnews"
	Readnum, err4 := httpget(url2) //点赞数
	if err4 != nil {
		err = err4
	}
	ari.ReadNum = Readnum[:6] //去空格

	//匹配正文
	re4 := regexp.MustCompile(`</h6>(?s:(.*?))" class="ar_article">`)
	joyUrls4 := re4.FindAllStringSubmatch(result, -1)

	var count int = 0 //
	for _, data := range joyUrls4 {
		count++
		if count <= 1 {
			continue
		}
		//正则匹配出来的字符串包含正文
		//这段为调试具体正文位置，也可以调试后直接写出，这边举例前面joyUrls的调试过程
		selector = "#" + data[1][51:]

		str, err4 = GetSpecialData(result, selector)
		if err4 != nil {
			err = err4
		}
		if count == 2 {
			break
		}
	}
	ari.Content = str
	return

}

func Spiderhtml(i int, page chan int) {

	url := "https://www.fzu.edu.cn/index/fdyw/" + strconv.Itoa(i) + ".htm"
	// strconv.Itoa((i-1)*50

	result, err := httpget(url)

	if err != nil {
		fmt.Println("http get err=", err)
		return
	}

	re := regexp.MustCompile(`<a href="(?s:(.*?))" target="_blank" title="`)

	if re == nil {
		fmt.Println("regexp.MustCompile err")
		return
	}

	joyUrls := re.FindAllStringSubmatch(result, -1)

	var count int = 0
	for _, data := range joyUrls {
		count++
		if count <= 2 {
			continue
		}

		str := data[1]
		ari, err := spider2(str)
		articles = append(articles, *ari)
		if err != nil {
			panic(err)
		}
	}
	page <- i

}
func DoWoke() {
	page := make(chan int)
	for i := 70; i >= 61; i-- {
		go Spiderhtml(i, page) //并发
	}

	for i := 70; i >= 61; i-- {
		<-page
	}

}

func main() {

	DoWoke()
	db, err := gorm.Open("mysql", "root:111111@(127.0.0.1:3306)/txxt?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 创建表 自动迁移 (把结构体和数据表进行对应)
	db.AutoMigrate(&Arti{})
	for _, v := range articles {

		db.Create(&v)
	}
}
