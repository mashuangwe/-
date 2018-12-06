package music

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/bitly/go-simplejson"
	//	"github.com/PuerkitoBio/goquery"
)

const (
	//	URL = "http://music.163.com/weapi/cloudsearch/get/web?csrf_token="
	//	URL = "http://music.163.com/weapi/search/get"
	URL      = "http://music.163.com/weapi/cloudsearch/get/web"
	SingsUrl = "http://music.163.com/song/media/outer/url?id="
)

func GenerateDeplayUrl(songId string) string {
	outDeployUrl := SingsUrl + songId + ".mp3"
	return outDeployUrl
}

func AskAPI(query string) ([]byte, bool) {
	stype := "1"
	offset := "0"
	limit := "9"
	preParams := "{\"s\": \"" + query + "\", \"type\": \"" + stype + "\", \"offset\": " + offset + ", \"limit\": " + limit + ", \"total\": true, \"csrf_token\": \"\"}"
	params, encSecKey, encErr := EncParams(preParams)
	if encErr != nil {
		beego.Debug(encErr)
		return []byte{}, false
	}

	client := &http.Client{}
	form := url.Values{}
	form.Set("params", params)
	form.Set("encSecKey", encSecKey)
	body := strings.NewReader(form.Encode())
	req, _ := http.NewRequest("POST", URL, body)

	// 头设置
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://music.163.com")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))
	req.Header.Set("User-Agent", fakeAgent())

	// 发起请求
	response, reqErr := client.Do(req)
	if reqErr != nil {
		beego.Debug(reqErr)
		return []byte{}, false
	}
	defer response.Body.Close()

	resBody, resErr := ioutil.ReadAll(response.Body)
	if resErr != nil {
		beego.Debug(resErr)
		return []byte{}, false
	}
	return resBody, true
}

func Wangyi(keywords, singer string, offset int) ([]map[string]string, bool) {
	var results []map[string]string
	query := fmt.Sprintf("%s %s", keywords, singer)
	body, ok := AskAPI(query)
	if !ok {
		return results, false
	}

	js, err := simplejson.NewJson(body)
	if err != nil {
		beego.Error("error when convert []byte to json. error: ", err)
		return results, false
	}

	data, ok := js.CheckGet("result")

	if !ok {
		beego.Error("do not have a tag: result")
		return results, false
	}

	song_count, err := data.Get("songCount").Int()
	if err != nil || song_count == 0 {
		beego.Error("do not hava a song.", err)
		return results, false
	}

	//	beego.Debug("song_count=", song_count)

	songs := data.Get("songs")
	sngs_lst, _ := songs.Array()

	for index, _ := range sngs_lst {
		tmp := songs.GetIndex(index)
		songname, _ := tmp.Get("name").String()
		songId, _ := tmp.Get("id").Int64()
		copyright, _ := tmp.Get("privilege").Get("cp").Int64() // 0-无版权，1-有版权

		//		beego.Debug("songname=", songname)
		//		beego.Debug("songId=", songId)
		//		beego.Debug("copyright=", copyright)

		songIdStr := strconv.FormatInt(songId, 10)
		mp3_url := GenerateDeplayUrl(songIdStr)
		pic_url, _ := tmp.Get("al").Get("picUrl").String()
		//		beego.Debug("pic_url=", pic_url)
		//		beego.Debug("mp3_url=", mp3_url)

		artists := tmp.Get("ar")
		artists_lst, _ := artists.Array()
		var singers []string

		for i, _ := range artists_lst {
			a_tmp := artists.GetIndex(i)
			singer, _ := a_tmp.Get("name").String()
			singers = append(singers, singer)
		}
		singer := strings.Join(singers, ",")
		song_detail := map[string]string{
			"song":      songname,
			"url":       mp3_url,
			"picurl":    pic_url,
			"singer":    singer,
			"copyright": strconv.FormatInt(copyright, 10),
		}

		results = append(results, song_detail)
	}

	return results, true
}
