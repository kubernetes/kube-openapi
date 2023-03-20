package common

import (
        "fmt"
        "io/ioutil"
        "net/http"
		"net/url"
        "html"
        "regexp"
		"time"
)

const GOOGLE_TRANSLATE_URL = "http://translate.google.com/m?tl=zh-CN&sl=en&q="

func TranslateDescription(sourceContent string) string {
	// 内容转换为URLCode
	Content := url.QueryEscape(sourceContent)
	url := GOOGLE_TRANSLATE_URL + Content
	// fmt.Printf(url)

	// 保证访问间隔0.2s
	time.Sleep(200 * time.Microsecond)
	response, err := http.Get(url)
	if err != nil {
			fmt.Println("请求错误", err.Error())
			// 保守翻译，遇错返回原值
			return sourceContent
	}
	// fmt.Println(response)

	// 解析Body结构中的数据
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
			fmt.Println("Body 读取错误", err)
			return sourceContent
	}
	// fmt.Println(string(data))

	// 正则过滤除翻译内容的其他信息
	expr := regexp.MustCompile(`(?s)class="(?:t0|result-container)">(.*?)<`)
	result := expr.FindAllStringSubmatch(string(data), -1)
	if len(result) == 0 {
			return sourceContent
	}

	targetContent := html.UnescapeString(result[0][1])
	// fmt.Println(targetContent)
	return targetContent
}