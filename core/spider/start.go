package spider

import (
	"github.com/go-redis/redis"
	"github.com/op/go-logging"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"myproj/try/testspider/model"
	"os"
	"fmt"
	"io"
	"time"
	"unicode"
	"strings"
	"path/filepath"
	"github.com/jinzhu/gorm"
	"myproj/try/testspider/config"
	"myproj/try/testspider/core/wxhandler"
	"gopkg.in/iconv.v1"
	"sync"
)

var client *redis.Client
var logger *logging.Logger

type ListToDown map[string]bool
var NianjiList = make(map[string]ListToDown)


var (
	DataRoot string
	RootUrl string
	listToDown ListToDown
)
func init()  {
	logger = logging.MustGetLogger("start")
	DataRoot = config.GetConfig().SavePath // 存放封面图的根目录
	RootUrl = config.GetConfig().RootUrl
	listToDown = ListToDown{}
	NianjiList = map[string]ListToDown{}
}

func checkPageUrl(sourcePageUrl string) error {
	db := model.CreateConn()
	tx := db.Begin()
	if err := tx.Debug().Model(&model.Question{}).Where("source_page_url = ?",sourcePageUrl).Find(&model.Question{}).Error;err != nil{
		if !gorm.IsRecordNotFoundError(err){
			return err
		}
		return nil
	}
	tx.Commit()

	return fmt.Errorf("record exist")
}

func getDoc(url string) *goquery.Document {
	resp,err := http.Get(url)
	if err != nil{
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Fatal(err)
	}
	return doc
}

func Start() error {
	//获取首页的html
	doc := getDoc(RootUrl)
	if doc == nil{
		return fmt.Errorf("get doc with error")
	}
	//获取一年级的链接
	yinianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(0).Find("a").Eq(0).Attr("href")
	//获取二年级的链接
	ernianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(1).Find("a").Eq(0).Attr("href")
	//获取三年级的链接
	sannianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(2).Find("a").Eq(0).Attr("href")
	//获取四年级的链接
	sinianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(3).Find("a").Eq(0).Attr("href")
	//获取五年级的链接
	wunianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(4).Find("a").Eq(0).Attr("href")
	//获取六年级的链接
	liunianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(5).Find("a").Eq(0).Attr("href")
	//
	logger.Info(yinianjiUrl,ernianjiUrl,sannianjiUrl,sinianjiUrl,wunianjiUrl,liunianjiUrl)
	handlenianji("liunianji",liunianjiUrl)

	handlenianji("wunianji",wunianjiUrl)
	handlenianji("yinianji",yinianjiUrl)
	handlenianji("ernianji",ernianjiUrl)
	handlenianji("sannianji",sannianjiUrl)
	handlenianji("sinianji",sinianjiUrl)


	//handler(config.GetConfig().TestPageUrl)

	return nil
}
var lock sync.Mutex
//var count int
var countmap = make(map[string]int)
func handlenianji(nianji string,nianjiurl string)  {
	var err error

	//count = 0
	NianjiList[nianji] = ListToDown{}
	countmap[nianji]=0

	lock.Lock()
	doc6 := getDoc(nianjiurl)

	//从六年级的首页中添加待下载的页面list
	doc6.Find("div.ttl-con").Find("ul").Find("a").Each(func(i int, content *goquery.Selection) {
		a, _ := content.Attr("href")
		if ok := NianjiList[nianji][a]; !ok{
			NianjiList[nianji][a] = false
		}

	})
	//下载count次问题
	for k,_ := range NianjiList[nianji]{

		if countmap[nianji]>0{
			break
		}
		//处理函数
		err = handler(k)
		if err != nil{
			logger.Info("handle with error")
			continue
		}
		countmap[nianji] ++
	}
	lock.Unlock()
}

type Pid struct {
	ID string
}




func handler(sourcePageUrl string) error {
	var err error
	//检查数据库是否已下载
	err = checkPageUrl(sourcePageUrl)
	if err != nil{
		logger.Info(err)
		return fmt.Errorf("maybe page already used")
	}
	logger.Info("make a new")
	//获取html 问题第一页
	resp,err := http.Get(sourcePageUrl)
	if err != nil{
		logger.Info(err)
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Fatal(err)
		return err
	}
	var cont,ansCont  string
	//获得问题的图片
	que,ansPageUrl,grade,queType := getQuePic(doc)
	logger.Info("get page 2")
	//获取html 问题第二页
	resp,err = http.Get(ansPageUrl)
	if err != nil{
		logger.Info(err)

		return err
	}
	defer resp.Body.Close()
	doc1, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Fatal(err)
		return err
	}
	//获取答案的图片
	ans := getAnsPic(doc1)
	//添加处理记录
	err = model.SaveToDB(que,ans,grade,queType,sourcePageUrl,cont,ansCont)
	if err != nil{
		logger.Info(err)

		return err
	}
	//没有图片时立刻返回，不进行提问和回答操作
	if que == ""||ans == ""{
		//cont = getContent(doc)
		//ansCont = getAnsContent(doc1)
		//logger.Info(cont,ansCont)
		return fmt.Errorf("no pic")
	}

	logger.Info(que,ans)

	wp := &Wallpaper{QueUrl:que,ansPageUrl:ans,Time:time.Now(),Grade:grade,queType:queType,SourcePageUrl:sourcePageUrl}
	//保存图片并提问和回答
	SaveImage(wp)


	return nil
}



type Wallpaper struct {
	Time   time.Time
	Grade   string
	queType string
	QueUrl   string
	ansPageUrl string
	SourcePageUrl string
}

func getQuePic(doc *goquery.Document) (string,string,string,string) {
	// Find the review items
	//logger.Info("-------------------")
	//a,_ := doc.GBKHtml()
	//logger.Info(a)
	//logger.Info("-------------------")
	//cont,_:= doc.Find("div.content").GBKHtml()
	//logger.Info(cont)
	//logger.Info("-------------------")
	cont,_:= doc.Find("div.content").Find("p").Eq(1).GBKHtml()
	logger.Info(cont)

	//获取该页图片
	res1,_ :=doc.Find("div.content").Find("img").Attr("src")
	//logger.Info(res1)
	//获取第二页链接
	res2,_ :=doc.Find("div.btn-pages").Find("a").Eq(1).Attr("href")
	//获取年级和类型
	res3,_ := doc.Find("div.content").Find("p").Eq(0).Find("strong").GBKHtml()

	aFunc := func(a rune) bool { return !unicode.IsLetter(a) }
	res := strings.FieldsFunc(res3, aFunc)
	logger.Info(res)
	if len(res)==0{
		return "","","",""
	}
	logger.Info(res[0][:9],res[len(res)-1])

	return res1,res2,res[0][:9],res[len(res)-1]
}

func getContent(doc *goquery.Document) (string) {
	cont,_ := doc.Find("div.content").Find("p").Eq(1).GBKHtml()
	//logger.Info(cont)
	return cont
}

func getAnsContent(doc *goquery.Document) (string) {
	cont,_ := doc.Find("div.content").Find("p").Eq(1).GBKHtml()
	aConv, err := iconv.Open("utf-8", "GBK")
	if err != nil{
		return ""
	}
	logger.Info(aConv.ConvString(string(cont)))
	logger.Info(cont)
	return cont
}

func getAnsPic(doc *goquery.Document) (string) {
	res1,_ :=doc.Find("div.content").Find("img").Attr("src")
	return res1
}


func SaveImage(paper *Wallpaper) {

	//按年级+时间目录保存图片
	Dirname := DataRoot + paper.Grade + "_" + paper.Time.Format("20060102") + "/"
	if ! isDirExist(Dirname) {
		os.Mkdir(Dirname, 0755);
		fmt.Printf("dir %s created\n", Dirname)
	}
	//下载问题图片
	picpath,err := download(paper.QueUrl,Dirname)
	if err != nil{
		return
	}
	//提交问题
	pid,err := wxhandler.SubmitQues(config.GetConfig().XcxUrl+"ask/",picpath,"test",paper.queType,paper.Grade)
	if err != nil{
		return
	}
	//下载答案图片
	picpath,err = download(paper.ansPageUrl,Dirname)
	if err != nil{
		return
	}
	//回答问题
	_,err = wxhandler.SubmitAns(config.GetConfig().XcxUrl+"submitanswer/",picpath,"test", pid,"oowmZ5R311ND3StOd4KBOUiT-XJI","答案如图","true")
	if err != nil{
		return
	}
}

func download(url string,dirname string) (string,error) {
	var err error
	//根据URL文件名创建文件
	filename := filepath.Base(url)
	dst, err := os.Create(dirname + filename)
	if err != nil {
		return "",err
	}
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return "",err
	}
	// 写入文件
	logger.Info("this is dis:",dirname + filename)
	io.Copy(dst, res.Body)
	return dirname+filename,nil
}


func isDirExist(path string) bool {
	p, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return p.IsDir()
	}
}

