package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cangyan/notion-qiita/types/date"
	"github.com/cangyan/notion-qiita/types/files"
	"github.com/cangyan/notion-qiita/types/filter_object"
	"github.com/cangyan/notion-qiita/types/multi_select"
	"github.com/cangyan/notion-qiita/types/number"
	"github.com/cangyan/notion-qiita/types/rich_text"
	"github.com/cangyan/notion-qiita/types/title"
	"github.com/cangyan/notion-qiita/types/url"
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

type NotionPageData struct {
	Parent struct {
		DatabaseId string `json:"database_id,omitempty"`
	} `json:"parent,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Children   interface{}            `json:"children,omitempty"`
}

func (r Resp) GetIds() []string {
	var ret []string
	for _, item := range r {
		ret = append(ret, item.Id)
	}

	return ret
}

func (r *RespItem) ToNotionPageData(dbId string) string {
	var ret NotionPageData
	ret.Parent.DatabaseId = dbId
	ret.Properties = make(map[string]interface{})
	// id
	{
		ret.Properties["ID"] = rich_text.ValueOf(r.Id)
		ret.Properties["标题"] = title.ValueOf(r.Title)
		ret.Properties["URL地址"] = url.ValueOf(r.Url)
		var tags []string
		for _, item := range r.Tags {
			tags = append(tags, item.Name)
		}
		ret.Properties["标签"] = multi_select.ValueOf(tags)
		ret.Properties["点赞数"] = number.ValueOf(float64(r.LikesCount))
		ret.Properties["作者名字"] = rich_text.ValueOf(r.User.Id)
		ret.Properties["作者头像"] = files.ValueOf(r.User.ProfileImageUrl)
		ret.Properties["创建日期"] = date.ValueOf(r.CreatedAt)
	}

	b, _ := json.Marshal(ret)
	return string(b)
}

func main() {
	page := 1
	perPage := 10
	days := 7
	stocks := 20
	var created string

	qiitaToken := os.Getenv("QIITA_TOKEN")
	notionToken := os.Getenv("NOTION_TOKEN")
	notionQiitaDb := os.Getenv("NOTION_QIITA_DB")
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

	fmt.Print(page, " ", perPage, " ", stocks, " ", created, " ", qiitaToken)
	fmt.Println("")
	qiitaResp, err := GetArticles(page, perPage, stocks, created, qiitaToken)
	if err != nil {
		return
	}

	// fmt.Println(err, resp)
	if notionToken == "" {
		err := fmt.Errorf("notion token not config")
		fmt.Println(err)
		return
	}

	if notionQiitaDb == "" {
		err := fmt.Errorf("notion qiita db not config")
		fmt.Println(err)
		return
	}

	existedIds := FindNotionPageExistedById(notionToken, notionQiitaDb, qiitaResp.GetIds())

	// fmt.Println(existedIds)
	// fmt.Println(len(existedIds))

	for _, item := range qiitaResp {
		// fmt.Println(item.ToNotionPageData(notionQiitaDb))
		if StringInArr(item.Id, existedIds) {
			fmt.Println(item.Id + "文章已存在")
			continue
		}
		err := CreateNotionPage(notionToken, item.ToNotionPageData(notionQiitaDb))
		if err != nil {
			fmt.Println(err)
		}
	}

	// fmt.Println(qiitaResp.GetIds())
}

func StringInArr(s string, arr []string) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}

	return false
}

func GetArticles(page, perPage, stocks int, created string, qiitaToken string) (Resp, error) {
	url := fmt.Sprintf("https://qiita.com/api/v2/items?page=%d&per_page=%d&query=created:>%s+stocks:>%d", page, perPage, created, stocks)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if qiitaToken != "" {
		req.Header.Add("Authorization", "Bearer "+qiitaToken)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp := Resp{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return resp, nil
}

func CreateNotionPage(token string, in string) error {
	url := "https://api.notion.com/v1/pages"
	method := "POST"

	payload := strings.NewReader(in)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("Notion-Version", "2021-08-16")
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// fmt.Println(string(body))

	return nil
}

type NotionDatabaseQueryResp struct {
	Object     string      `json:"object,omitempty"`
	Results    []Block     `json:"results,omitempty"`
	NextCursor interface{} `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more,omitempty"`
}

type Block struct {
	Object     string
	Id         string
	Properties struct {
		ID struct {
			Title []struct {
				Text struct {
					Content string
				}
			}
		}
	}
}

func (r *NotionDatabaseQueryResp) GetArticleIds() []string {
	data := make([]string, 0)
	if len(r.Results) > 0 {
		for _, item := range r.Results {
			data = append(data, item.Properties.ID.Title[0].Text.Content)
		}
	}

	return data
}

func FindNotionPageExistedById(token string, dbId string, aIds []string) []string {
	ret := make([]string, 0)
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%v/query", dbId)
	method := "POST"

	payload := strings.NewReader(`{"filter":` + filter_object.GenerateTextOrFilterObject("ID", aIds) + `}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return ret
	}
	req.Header.Add("Notion-Version", "2021-08-16")
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ret
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ret
	}

	// fmt.Println(string(body))
	resp := NotionDatabaseQueryResp{}

	_ = json.Unmarshal(body, &resp)

	return resp.GetArticleIds()
}
