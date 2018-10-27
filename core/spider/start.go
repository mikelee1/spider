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
	"io/ioutil"
	"bytes"
	"mime/multipart"
)

var client *redis.Client
var logger *logging.Logger

var (
	DataRoot string
	RootUrl string
	ListToDown map[string]bool
)
func init()  {
	logger = logging.MustGetLogger("start")
	DataRoot = config.GetConfig().SavePath // 存放封面图的根目录
	RootUrl = config.GetConfig().RootUrl
	ListToDown = make(map[string]bool)
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
	yinianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Find("a").Eq(0).Attr("href")
	//获取二年级的链接
	ernianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(1).Find("a").Eq(0).Attr("href")
	//获取六年级的链接
	sannianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(2).Find("a").Eq(0).Attr("href")
	//获取六年级的链接
	sinianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(3).Find("a").Eq(0).Attr("href")
	//获取六年级的链接
	wunianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(4).Find("a").Eq(0).Attr("href")
	//获取六年级的链接
	liunianjiUrl,_ := doc.Find("div.tp10").Find("h2.col-tit").Eq(5).Find("a").Eq(0).Attr("href")
	//
	handlenianji(yinianjiUrl)
	handlenianji(ernianjiUrl)
	handlenianji(sannianjiUrl)
	handlenianji(sinianjiUrl)
	handlenianji(wunianjiUrl)
	handlenianji(liunianjiUrl)

	return nil
}

func handlenianji(nianji string)  {
	var err error
	var count int
	doc6 := getDoc(nianji)

	//logger.Info(doc6.GBKHtml())
	//doc6.Find("div.calendar_cn").Find("tbody").Find("a").Each(func(i int, content *goquery.Selection) {
	//	a, _ := content.Attr("href")
	//	if ok := ListToDown[a]; !ok{
	//		ListToDown[a] = false
	//	}else{
	//		count ++
	//	}
	//
	//})
	//从六年级的首页中添加待下载的页面list
	doc6.Find("div.ttl-con").Find("ul").Find("a").Each(func(i int, content *goquery.Selection) {
		a, _ := content.Attr("href")
		if ok := ListToDown[a]; !ok{
			ListToDown[a] = false
		}else{
			count ++
		}

	})
	logger.Info(ListToDown)
	logger.Info(count)
	//下载count次问题
	for k,_ := range ListToDown{

		if count>1{
			break
		}
		//处理函数
		err = handler(k)
		if err != nil{
			continue
		}
		count ++
	}
}

type Pid struct {
	ID string
}

func SubmitQues(url, filePath,filename,desc,grade string) (string,error) {

	//打开文件句柄操作
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return "",err
	}
	defer file.Close()

	//创建一个模拟的form中的一个选项,这个form项现在是空的
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作, 设置文件的上传参数叫uploadfile, 文件名是filename,
	//相当于现在还没选择文件, form项里选择文件的选项
	fileWriter, err := bodyWriter.CreateFormFile("problempic", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return "",err
	}

	//iocopy 这里相当于选择了文件,将文件放到form中
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return "",err
	}

	//获取上传文件的类型,multipart/form-data; boundary=...
	contentType := bodyWriter.FormDataContentType()

	//这个很关键,必须这样写关闭,不能使用defer关闭,不然会导致错误
	bodyWriter.Close()


	//这里就是上传的其他参数设置,可以使用 bodyWriter.WriteField(key, val) 方法
	//也可以自己在重新使用  multipart.NewWriter 重新建立一项,这个再server 会有例子
	params := map[string]string{
		"filename":filename,
		"userid":"oowmZ5a-8wr0A4Dg5pBHthCnY3vc",
		"desc":desc,
		"avatar":"https://mathoj.liyuanye.club/static/avatar/oowmZ5a-8wr0A4Dg5pBHthCnY3vc.jpg",
		"grade":grade,
		"easy":"简单",
		"reward":"1个奥币",
		//"problempic":,
	}
	//这种设置值得仿佛 和下面再从新创建一个的一样
	for key, val := range params {
		_ = bodyWriter.WriteField(key, val)
	}

	//发送post请求到服务端
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return "",err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "",err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))


	return string(resp_body),nil
}


//imgsign="true"
func SubmitAns(url, filePath,filename,problemid,teacherid,textsolu,imgsign string) (int,error) {

	//打开文件句柄操作
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return 0,err
	}
	defer file.Close()

	//创建一个模拟的form中的一个选项,这个form项现在是空的
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作, 设置文件的上传参数叫uploadfile, 文件名是filename,
	//相当于现在还没选择文件, form项里选择文件的选项
	fileWriter, err := bodyWriter.CreateFormFile("image", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return 0,err
	}

	//iocopy 这里相当于选择了文件,将文件放到form中
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return 0,err
	}

	//获取上传文件的类型,multipart/form-data; boundary=...
	contentType := bodyWriter.FormDataContentType()

	//这个很关键,必须这样写关闭,不能使用defer关闭,不然会导致错误
	bodyWriter.Close()


	//这里就是上传的其他参数设置,可以使用 bodyWriter.WriteField(key, val) 方法
	//也可以自己在重新使用  multipart.NewWriter 重新建立一项,这个再server 会有例子
	logger.Info("problemid is: ",problemid)
	params := map[string]string{

		"problemid":problemid,
		"teacherid":teacherid,
		"textsolu":textsolu,
		"imgsign":imgsign,

		"filename":filename,
	}
	//这种设置值得仿佛 和下面再从新创建一个的一样
	for key, val := range params {
		_ = bodyWriter.WriteField(key, val)
	}

	//发送post请求到服务端
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return 0,err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0,err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	//tmpPid := &Pid{}
	//json.Unmarshal(resp_body,tmpPid)
	return 0,nil
}


func handler(sourcePageUrl string) error {
	var err error
	//检查数据库是否已下载
	err = checkPageUrl(sourcePageUrl)
	if err != nil{
		return fmt.Errorf("maybe page already used")
	}
	logger.Info("make a new")
	//获取html 问题第一页
	resp,err := http.Get(sourcePageUrl)
	if err != nil{
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Fatal(err)
	}
	var cont,ansCont  string
	//获得问题的图片
	que,ansPageUrl,grade,queType := getQuePic(doc)

	//获取html 问题第二页
	resp,err = http.Get(ansPageUrl)
	if err != nil{
		return err
	}
	defer resp.Body.Close()
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Fatal(err)
	}
	//获取答案的图片
	ans := getAnsPic(doc)
	//添加处理记录
	err = SaveToDB(que,ans,grade,queType,sourcePageUrl,cont,ansCont)
	if err != nil{
		return err
	}
	//没有图片时立刻返回，不进行提问和回答操作
	if que == ""||ans == ""{
		cont = getContent(doc)
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
	//cont,_= doc.Find("div.content").Find("p").Eq(1).GBKHtml()
	//logger.Info(cont)

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
	pid,err := SubmitQues(config.GetConfig().XcxUrl+"ask/",picpath,"test",paper.queType,paper.Grade)
	if err != nil{
		return
	}
	//下载答案图片
	picpath,err = download(paper.ansPageUrl,Dirname)
	if err != nil{
		return
	}
	//回答问题
	_,err = SubmitAns(config.GetConfig().XcxUrl+"submitanswer/",picpath,"test", pid,"oowmZ5R311ND3StOd4KBOUiT-XJI","答案如图","true")
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

func SaveToDB(qus,ans,grade,queType,sourcePageUrl,content,ansContent string) error {
	var err error
	db := model.CreateConn()
	tx := db.Begin()
	tmp := &model.Question{QuesPic:qus,AnsPic:ans,Grade:grade,QuesType:queType,SourcePageUrl:sourcePageUrl,Content:content,AnsContent:ansContent}

	if err = tx.Model(&model.Question{}).Save(tmp).Error;err != nil{
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}