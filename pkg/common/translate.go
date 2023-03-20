package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const YOUDAO_TRANSLATE_URL = "http://fanyi.youdao.com"

func TranslateDescription(sourceContent string) string {
	// 内容转换为URLCode
	Content := url.QueryEscape(sourceContent)
	// 定义url，例如：http://fanyi.youdao.com/translate?&doctype=json&type=EN2ZH_CN&i=hello
	url := YOUDAO_TRANSLATE_URL + "/translate?&doctype=text&type=EN2ZH_CN&i=" + Content

	// 获取http请求
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("请求错误", err.Error())
		// 错误返回源内容
		return sourceContent
	}

	// 获取内容
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Body 读取错误", err.Error())
		// 错误返回源内容
		return sourceContent
	}

	// 排除长度为0
	if len(data) == 0 {
		return sourceContent
	}

	// 解析text内容，提取翻译内容，例如：
	// errorCode=0
	// result=你好
	// 匹配结果：你好
	m := regexp.MustCompile(`[^errorCode=0][^result=*?](.*)`)
	result := m.FindString(string(data))
	return result
}
