package wxhandler

import (
	"os"
	"fmt"
	"bytes"
	"mime/multipart"
	"io"
	"net/http"
	"io/ioutil"
	"github.com/op/go-logging"
)

var logger *logging.Logger

func init()  {
	logger = logging.MustGetLogger("wxhandler")
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
