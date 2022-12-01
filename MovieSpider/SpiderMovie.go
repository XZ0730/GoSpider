package main

/*
	1.动态数据爬取
	————Edge chorme 工具运用
	————通过查找对应数据的api的url获取数据
	2.熟悉掌握正则匹配的运用
	————导regexp正则库
		——regexp中函数MustCompile  FindAllStringSubmatch的运用
		——正则表达式
	3。gorm一对多关系映射
	————父级表要又子级表的结构体集合，名称为子级表后加s，默认绑定id字段，
	可以通过加tag{`gorm:"forignKey:子表属性"`}选择关联的子表属性值
	————gorm启动数据库，CRUD操作

*/
import (
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type M_Comment struct {
	gorm.Model
	MajorComment string `gorm:"type:text"`

	S_Comments []S_Comment `gorm:"foreignKey:Mid"`
}
type S_Comment struct {
	gorm.Model
	SecondComment string `gorm:"type:longtext"`
	Mid           int64
	// Isempty       int    `gorm:"default:1"`
}

// dataBase , userName ,passWord ,IP ,Port ,dbName
const (
	dataBase string = "mysql"
	userName string = "root"
	passWord string = "111111"
	IP       string = "127.0.0.1"
	Port     string = "3306"
	dbName   string = "bilibliComment"
)

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
func Regexpp(html string, target string) [][]string {
	re := regexp.MustCompile(`target`)
	str := re.FindAllStringSubmatch(html, -1)
	return str
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

var count int = 0   //没什么用，用来调试的时候输出子评论的索引，直接删了
var pre string = "" //前一个root id
var count1 int = 0  //限制爬取主评论数
var index int = 1   //一个main api中评论的索引

func Spiderhtml() {
	var Comment_temp M_Comment
	var Stemp S_Comment
	//参数信息 全局：[]M_Comment 局部

	for i := 1; ; i++ {
		result2, _ := httpget("https://api.bilibili.com/x/v2/reply/main?jsonp=jsonp&next=" + strconv.Itoa(i) + "&type=1&oid=21071819&mode=3&plat=1&_=")

		re := regexp.MustCompile(`"is_end":(?s:(.*?)),"`) //主评论是否结束
		is_end := re.FindAllStringSubmatch(result2, -1)
		fmt.Println(is_end[0][1])
		if is_end[0][1] == "false" {
			re1 := regexp.MustCompile(`root":(?s:(.*?)),"`)
			rpid := re1.FindAllStringSubmatch(result2, -1)

			re5 := regexp.MustCompile(`"message":"(?s:(.*?))",`)
			mcomment := re5.FindAllStringSubmatch(result2, -1)
			fmt.Println(rpid) //这个是root id  用于拼接子评论api的url

			index = 1
			for _, rid := range rpid { //主评论id  根据这个id访问子评论api
				if rid[1] == "0" { //遍历到root id为 0 的时候就是主评论 否则不是
					fmt.Println("这是主评论", mcomment[index][1])
					Comment_temp.MajorComment = mcomment[index][1]
					count++
					index++
					continue
				}
				if rid[1] == pre { //去重
					index++
					continue
				}
				count1++
				index++
				fmt.Println(rid[1])
				time.Sleep(2 * time.Second) //调试用的，防止爬取太快看不到输出
				pre = rid[1]
				// fmt.Println(rid[1])
				result, err := httpget("https://api.bilibili.com/x/v2/reply/reply?jsonp=jsonp&pn=1&type=1&oid=21071819&ps=10&root=" + rid[1] + "&_=1669794927981")
				if err != nil {
					panic(err)
				}
				re2 := regexp.MustCompile(`"sub_reply_entry_text":"共(?s:(.*?))条回复"`) //子评论数
				replyNum := re2.FindAllStringSubmatch(result, -1)
				fmt.Println(replyNum)
				reply1, err1 := strconv.Atoi(replyNum[0][1])
				reply1 = int(math.Ceil(float64(reply1) / 10)) //子评论的页数  向上取整
				count = 0
				for j := 1; j <= reply1; j++ { //
					result1, err := httpget("https://api.bilibili.com/x/v2/reply/reply?jsonp=jsonp&pn=" + strconv.Itoa(j) + "&type=1&oid=21071819&ps=10&root=" + rid[1] + "&_=")
					re3 := regexp.MustCompile(`"message":"(?s:(.*?))",`) //爬取 reply 中的所有评论
					res := re3.FindAllStringSubmatch(result1, -1)

					// fmt.Println(res)
					for _, v := range res { //爬取所有评论
						if v[1] == "0" || v[1] == Comment_temp.MajorComment {
							continue //rply api每一页都包含主评论，这边要去重
						}
						Stemp.SecondComment = v[1] //子评论赋值
						Comment_temp.S_Comments = append(Comment_temp.S_Comments, Stemp)

					}
					num := len(Comment_temp.S_Comments)
					Comment_temp.S_Comments = Comment_temp.S_Comments[:num-1]

					// fmt.Println(len(Comment_temp.S_Comments))
					if err != nil {
						panic(err)
					}

				}
				M_comments = append(M_comments, Comment_temp)

				if err1 != nil {
					panic(err1)

				}
				if count1 == 3 { //这边count1是代表只爬三条主评论方便展示数据结果，因为爬太多会触发b站反爬机制，这个可以和下面的count1条件一起去掉就可以全爬
					break
				}

			}
			if count1 == 3 {
				break
			}
		} else {
			break
		}
	}
}
func DoWoke() {

	Spiderhtml()

}

func main() {
	DoWoke()
	//dataBase , userName ,passWord ,IP ,Port ,dbName
	db1, err := gorm.Open(dataBase, userName+":"+passWord+"@("+IP+":"+Port+")/"+dbName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	// 创建表 自动迁移 (把结构体和数据表进行对应)
	db1.AutoMigrate(&M_Comment{}, &S_Comment{})

	for _, v := range M_comments {
		db1.Create(&v)
	}

}
