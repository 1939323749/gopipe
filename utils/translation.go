package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/abadojack/whatlanggo"
	"github.com/andybalholm/brotli"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Lang struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"`
	TargetLang             string `json:"target_lang"`
}
type CommonJobParams struct {
	WasSpoken    bool   `json:"wasSpoken"`
	TranscribeAS string `json:"transcribe_as"`
}
type Text struct {
	Text                string `json:"text"`
	RequestAlternatives int    `json:"requestAlternatives"`
}
type Params struct {
	Texts           []Text          `json:"texts"`
	Splitting       string          `json:"splitting"`
	Lang            Lang            `json:"lang"`
	Timestamp       int64           `json:"timestamp"`
	CommonJobParams CommonJobParams `json:"commonJobParams"`
}
type PostData struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int64  `json:"id"`
	Params  Params `json:"params"`
}

func initData(sourceLang string, targetLang string) *PostData {
	return &PostData{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		Params: Params{
			Splitting: "newlines",
			Lang: Lang{
				SourceLangUserSelected: sourceLang,
				TargetLang:             targetLang,
			},
			CommonJobParams: CommonJobParams{
				WasSpoken:    false,
				TranscribeAS: "",
				// RegionalVariant: "en-US",
			},
		},
	}
}
func getICount(translateText string) int64 {
	return int64(strings.Count(translateText, "i"))
}

func getRandomNumber() int64 {
	num := rand.Int63n(99999) + 8300000
	return num * 1000
}

func getTimeStamp(iCount int64) int64 {
	ts := time.Now().UnixMilli()
	if iCount != 0 {
		iCount = iCount + 1
		return ts - ts%iCount + iCount
	} else {
		return ts
	}
}

func Translation(sourceLang string, targetLang string, translateText string) (string, error) {
	id := getRandomNumber()
	url := "https://www2.deepl.com/jsonrpc"
	if sourceLang == "" {
		lang := whatlanggo.DetectLang(translateText)
		deepLLang := strings.ToUpper(lang.Iso6391())
		sourceLang = deepLLang
	}
	if targetLang == "" {
		targetLang = "ZH"
	}
	text := Text{
		Text:                translateText,
		RequestAlternatives: 3,
	}

	//debug
	//fmt.Print(sourceLang, targetLang, translateText)

	postData := initData(sourceLang, targetLang)
	// set id
	postData.ID = id
	// set text
	postData.Params.Texts = append(postData.Params.Texts, text)
	// set timestamp
	postData.Params.Timestamp = getTimeStamp(getICount(translateText))
	postByte, _ := json.Marshal(postData)
	postStr := string(postByte)

	// add space if necessary
	if (id+5)%29 == 0 || (id+3)%13 == 0 {
		postStr = strings.Replace(postStr, "\"method\":\"", "\"method\" : \"", -1)
	} else {
		postStr = strings.Replace(postStr, "\"method\":\"", "\"method\": \"", -1)
	}

	postByte = []byte(postStr)
	reader := bytes.NewReader(postByte)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Println(err)
		return "", nil
	}

	// Set Headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "*/*")
	request.Header.Set("x-app-os-name", "iOS")
	request.Header.Set("x-app-os-version", "16.3.0")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("x-app-device", "iPhone13,2")
	request.Header.Set("User-Agent", "DeepL-iOS/2.9.1 iOS 16.3.0 (iPhone13,2)")
	request.Header.Set("x-app-build", "510265")
	request.Header.Set("x-app-version", "2.9.2")
	request.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			os.Exit(1)
		}
	}(resp.Body)

	var bodyReader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		bodyReader = brotli.NewReader(resp.Body)
	default:
		bodyReader = resp.Body
	}

	body, err := io.ReadAll(bodyReader)

	res := gjson.ParseBytes(body)

	//debug
	//fmt.Print(res.Get("result.texts.0.text").String())

	if res.Get("error.code").String() == "-32600" {
		return "", fmt.Errorf("%s", res.Get("error.data.what").String())
	}
	if res.Get("error.code").String() == "1042912" || res.Get("error.code").String() == "1042911" {
		return "", fmt.Errorf("%s", res.Get("error.message").String())
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("%s", res.Get("error").String())
	}
	return res.Get("result.texts.0.text").String(), nil
}
