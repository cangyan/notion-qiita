package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Resp []RespItem

type RespItem struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
	Tags  []struct {
		Name string `json:"name,omitempty"`
	} `json:"tags,omitempty"`
	LikesCount int    `json:"likes_count,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	User       struct {
		Id              string `json:"id,omitempty"`
		ProfileImageUrl string `json:"profile_image_url,omitempty"`
	} `json:"user,omitempty"`
}

func main() {
	page := 1
	perPage := 20
	days := 7
	stocks := 20
	var created string

	qiitaToken := os.Getenv("QIITA_TOKEN")
	pageStr := os.Getenv("QIITA_PAGE")
	perPageStr := os.Getenv("QIITA_PERPAGE")
	daysStr := os.Getenv("QIITA_DAYS")
	stocksStr := os.Getenv("QIITA_STOCKS")

	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	if perPageStr != "" {
		perPage, _ = strconv.Atoi(perPageStr)
	}

	if daysStr != "" {
		days, _ = strconv.Atoi(daysStr)
	}
	created = time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	if stocksStr != "" {
		stocks, _ = strconv.Atoi(stocksStr)
	}

	url := fmt.Sprintf("https://qiita.com/api/v2/items?page=%d&per_page=%d&query=created:>%s+stocks:>%d", page, perPage, created, stocks)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	if qiitaToken != "" {
		req.Header.Add("Authorization", "Bearer "+qiitaToken)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp := Resp{}
	err = json.Unmarshal(body, &resp)
	fmt.Println(err, resp)
}
