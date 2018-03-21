package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	dealid          = "1111"
	accessKeyId     = "accessKeyId"
	accessKeySecret = "accessKeySecret"

	cookie  = "xxxxxxxxxxxx"
	method  = "POST"
	uri     = "http://rtq.audienx.com/RTQ/blk"
	urlpath = "/RTQ/blk"
)

func GetRandom() uint64 {
	source := rand.NewSource(time.Now().UnixNano())
	newRand := rand.New(source)
	return newRand.Uint64()
}

type Request struct {
	DealId string `json:"dealid"`
	Cookie string `json:"cookie"`
	//IP     string `json:"ip"`
	//Device string `json:"device"`

	/*
		//optional
		MediaName    string `json:"mediaName"`
		PlatformType string `json:"platformType"`
		Refer        string `json:"refer"`
		UserIP       string `json:"userIP"`
	*/
}

type RequestBody []Request

type Description struct {
	code  int64  `json:code`
	Desc  string `json:desc`
	Score int64  `json:score`
}

type Response struct {
	Msg    string        `json:msg`
	Result []Description `json:result`
}

func GetSignature(method string, uri string, nonce uint64, ts int64,
	requestBody RequestBody, accessKeySecret string) string {

	var buffer bytes.Buffer
	buffer.WriteString(method)
	buffer.WriteString(uri)
	buffer.WriteString(strconv.FormatUint(nonce, 10))
	buffer.WriteString(strconv.FormatInt(ts, 10))

	//buffer.WriteString("body=")

	body, err := json.Marshal(requestBody)
	if err != nil {
		return ""
	}

	buffer.Write(body)

	//hmac ,use sha1
	key := []byte(accessKeySecret)
	mac := hmac.New(sha1.New, key)
	mac.Write(buffer.Bytes())

	sum := mac.Sum(nil)
	base64sum := base64.StdEncoding.EncodeToString(sum)
	return base64sum
}

func main() {
	requestBody := []Request{}
	requestBody = append(requestBody,
		Request{
			DealId: dealid,
			Cookie: cookie,
		})

	nonce := GetRandom()
	timestamp := time.Now().Unix()
	signature := GetSignature(method, urlpath, nonce, timestamp, requestBody, accessKeySecret)

	Body, err := json.Marshal(requestBody)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(Body))
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("s", url.QueryEscape(signature))
	q.Add("a", accessKeyId)
	q.Add("n", strconv.FormatUint(nonce, 10))
	q.Add("t", strconv.FormatInt(timestamp, 10))
	q.Add("v", "1")
	q.Add("mv", "1")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var responseBody Response = Response{}
	err = json.Unmarshal(content, &responseBody)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(responseBody)
}
