package music

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"smartvoice/common"
	"time"

	"github.com/astaxie/beego"
)

type Music struct {
	m string
}

// 歌手作品列表
func (m *Music) SingerComposition(params []common.Parameter) string {
	agentid, sessionid, singer, _ := ConvList2Param(params)
	var resp common.Response
	resp.Service = "Music"
	resp.Operation = "SingerComposition"
	resp.Rc = "0"

	var data common.Text
	songs, ok := Wangyi("", singer, 0)
	if !ok {
		songs, _ = DBGetSongsBySinger(singer)
	}

	if len(songs) > 0 {
		data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
		var data_content [][]string
		for _, det := range songs {
			details := []string{det["song"], det["singer"], det["picurl"]}
			data_content = append(data_content, details)
		}
		data.DataContent = data_content
		resp.Rc = "0"
		resp.Hit = fmt.Sprintf("Find the following songs of %s for you:", singer)
		resp.TextContent = data
	} else {
		resp.Hit = fmt.Sprintf("Sorry, No song of %s was found", singer)
	}

	// save session
	redis_key := fmt.Sprintf("%s:%s:singer", agentid, sessionid)
	RedisSetKeyValue(redis_key, singer)
	message_string, _ := json.Marshal(resp)
	return string(message_string)
}

// 通过歌手查歌曲
func (m *Music) SongBySinger(params []common.Parameter) string {
	agentid, sessionid, singer, song := ConvList2Param(params)
	var resp common.Response
	resp.Service = "Music"
	resp.Operation = "SongBySinger"
	resp.Rc = "0"
	var data common.Text

	songs, ok := Wangyi(song, singer, 0)
	if !ok {
		songs, _ = DBGetSongsBySinger(singer)
	}

	//var song string
	if len(songs) > 0 {
		max := 60
		if len(songs) <= 60 {
			max = len(songs)
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		rindex := r.Intn(max)

		if songs[rindex]["copyright"] == "1" {
			song = songs[rindex]["song"]
			data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
			var data_content [][]string
			details := []string{songs[rindex]["song"], songs[rindex]["singer"], songs[rindex]["picurl"]}
			data_content = append(data_content, details)
			data.DataContent = data_content
			resp.TextContent = data
			resp.Hidden = songs[rindex]["url"]

			// save session
			redis_key := fmt.Sprintf("%s:%s:singer", agentid, sessionid)
			RedisSetKeyValue(redis_key, songs[rindex]["singer"])
			redis_key = fmt.Sprintf("%s:%s:song", agentid, sessionid)
			RedisSetKeyValue(redis_key, songs[rindex]["song"])
		} else {
			resp.Service = "qa"
			resp.Operation = ""
			resp.Hit = "我还不会唱这首歌哦"
		}
	} else {
		resp.Hit = fmt.Sprintf("抱歉,没有找到%s的作品", singer)
	}

	message_string, _ := json.Marshal(resp)
	return string(message_string)
}

// 通过歌曲名称和歌手搜索
func (m *Music) SongByNameAndSinger(params []common.Parameter) string {
	agentid, sessionid, singer, song := ConvList2Param(params)
	var resp common.Response
	resp.Service = "Music"
	resp.Operation = "SongByNameAndSinger"
	var data common.Text
	var song_dict map[string]string

	resp.Rc = "0"
	songs, ok := Wangyi(song, singer, 0)

	beego.Debug("song:", song, ",singer:", singer)
	beego.Debug("len(songs):", len(songs))
	if !ok || len(songs) == 0 {
		song_dict, _ = DBGetSongBySingerAndName(singer, song)
	} else {
		song_dict = songs[0]
	}

	if len(song_dict) > 0 {
		if song_dict["copyright"] == "1" {
			data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
			var data_content [][]string
			details := []string{song_dict["song"], song_dict["singer"], song_dict["picurl"]}
			data_content = append(data_content, details)
			data.DataContent = data_content

			resp.TextContent = data
			resp.Hidden = song_dict["url"]

			redis_key := fmt.Sprintf("%s:%s:song", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["song"])

			redis_key = fmt.Sprintf("%s:%s:singer", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["singer"])
		} else {
			resp.Service = "qa"
			resp.Operation = ""
			resp.Hit = "我还不会唱这首歌哦"
		}
	} else {
		// 后续可推荐该歌手的其他歌曲
		resp.Hit = fmt.Sprintf("抱歉,没有找到%s的%s", singer, song)
	}

	message_string, _ := json.Marshal(resp)
	return string(message_string)
}

// 通过歌曲名称搜索
func (m *Music) SongByName(params []common.Parameter) string {
	agentid, sessionid, singer, song := ConvList2Param(params)
	var resp common.Response
	var data common.Text
	resp.Service = "Music"
	resp.Operation = "SongByName"
	resp.Rc = "0"
	var song_dict map[string]string

	songs, ok := Wangyi(song, singer, 0)
	beego.Debug("song:", song)
	beego.Debug("len(songs):", len(songs))
	if !ok || len(songs) == 0 {
		song_dict, _ = DBGetSongByName(song)
	} else {
		song_dict = songs[0]
	}

	if len(song_dict) > 0 {
		if song_dict["copyright"] == "1" {
			//resp.Hit = fmt.Sprintf("现在为您播放%s的%s", song_dict["singer"], song)
			data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
			var data_content [][]string
			details := []string{song_dict["song"], song_dict["singer"], song_dict["picurl"]}
			data_content = append(data_content, details)
			data.DataContent = data_content

			resp.TextContent = data
			resp.Hidden = song_dict["url"]

			redis_key := fmt.Sprintf("%s:%s:song", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["song"])

			redis_key = fmt.Sprintf("%s:%s:singer", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["singer"])
		} else {
			resp.Service = "qa"
			resp.Operation = ""
			resp.Hit = "我还不会唱这首歌哦"
		}

	} else {
		// 后续可以推荐其他歌曲
		resp.Hit = fmt.Sprintf("Sorry, didn't find the song %s", song)
	}

	message_string, _ := json.Marshal(resp)
	return string(message_string)
}

// 随机给出一首歌
// query: 唱首歌
func (m *Music) SongRandomly(params []common.Parameter) string {
	agentid, sessionid, _, _ := ConvList2Param(params)
	var resp common.Response
	var data common.Text
	resp.Service = "Music"
	resp.Operation = "SongRandomly"
	resp.Rc = "0"
	var song_dict map[string]string

	songs, ok := Wangyi("英文歌曲", "", 0)
	beego.Debug("len(songs):", len(songs))
	if !ok || len(songs) == 0 {
		song_dict, _ = DBGetRandomOne()
	} else {
		max := 60
		if len(songs) <= 60 {
			max = len(songs)
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		rindex := r.Intn(max)
		song_dict = songs[rindex]
	}

	if len(song_dict) > 0 {
		if song_dict["copyright"] == "1" {
			//beego.Debug("++++++++", song_dict)
			//resp.Hit = fmt.Sprintf("现在为您播放%s的%s", song_dict["singer"], song_dict["song"])
			data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
			var data_content [][]string
			details := []string{song_dict["song"], song_dict["singer"], song_dict["picurl"]}
			data_content = append(data_content, details)
			data.DataContent = data_content

			resp.TextContent = data
			resp.Hidden = song_dict["url"]

			redis_key := fmt.Sprintf("%s:%s:singer", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["singer"])
			redis_key = fmt.Sprintf("%s:%s:song", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["song"])
		} else {
			resp.Service = "qa"
			resp.Operation = ""
			resp.Hit = "我还不会唱这首歌哦"
		}
	} else {
		// 连老歌都获取失败, 没办法,暂定进QA吧.好歹QA的回复能多样化一点
		resp.Rc = "-1"
	}

	message_string, _ := json.Marshal(resp)
	return string(message_string)
}

// 再唱一首歌
func (m *Music) SingAnthorSong(params []common.Parameter) string {
	agentid, sessionid, _, _ := ConvList2Param(params)
	var resp common.Response
	var data common.Text
	resp.Service = "Music"
	resp.Operation = "SingAnthorSong"

	var singer, song string
	//var err error
	var song_dict map[string]string

	redis_key := fmt.Sprintf("%s:%s:singer", agentid, sessionid)
	singer, _ = RedisGetValuebyKey(redis_key)
	redis_key = fmt.Sprintf("%s:%s:song", agentid, sessionid)
	song, _ = RedisGetValuebyKey(redis_key)

	if singer == "" {
		// 没有历史信息,从百度随便搜一个吧
		songs, ok := Wangyi("", "经典老歌", 0)
		if !ok || len(songs) == 0 {
			// 连老歌都获取失败,上数据库里随便拿一个
			song_dict, _ = DBGetRandomOne()
		} else {
			max := 60
			if len(songs) <= 60 {
				max = len(songs)
			}

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			rindex := r.Intn(max)
			song_dict = songs[rindex]
		}
	} else {
		// 在拿出一个这个singer的song来，和之前的这个song不一样就可以
		songs, ok := Wangyi("", singer, 0)
		if !ok {
			songs, ok = DBGetSongsBySinger(singer)
		}

		if !ok {
			// 没有找到这个歌手的歌曲，那随便拿一个出来好了
			song_dict, ok = DBGetRandomOne()
		}

		for _, song_details := range songs {
			if song_details["song"] != song {
				song_dict = song_details
				break
			}
		}
	}

	if len(song_dict) == 0 {
		resp.Hit = fmt.Sprintf("Sorry, I didn't find it.")
	} else {
		if song_dict["copyright"] == "1" {
			//resp.Hit = fmt.Sprintf("为您播放%s的%s", singer, song_dict["song"])
			data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
			var data_content [][]string
			details := []string{song_dict["song"], song_dict["singer"], song_dict["picurl"]}
			data_content = append(data_content, details)
			data.DataContent = data_content

			resp.TextContent = data
			resp.Hidden = song_dict["url"]

			redis_key := fmt.Sprintf("%s:%s:singer", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["singer"])
			redis_key = fmt.Sprintf("%s:%s:song", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["song"])
		} else {
			resp.Service = "qa"
			resp.Operation = ""
			resp.Hit = "我还不会唱这首歌哦"
		}
	}

	message_string, _ := json.Marshal(resp)
	return string(message_string)
}

func (m *Music) SingAgain(params []common.Parameter) string {
	agentid, sessionid, _, _ := ConvList2Param(params)
	var resp common.Response
	var data common.Text
	resp.Service = "Music"
	resp.Operation = "SingAnthorSong"

	var singer, song string
	//var err error
	var song_dict map[string]string

	redis_key := fmt.Sprintf("%s:%s:singer", agentid, sessionid)
	singer, _ = RedisGetValuebyKey(redis_key)
	redis_key = fmt.Sprintf("%s:%s:song", agentid, sessionid)
	song, _ = RedisGetValuebyKey(redis_key)

	songs, ok := Wangyi(song, singer, 0)
	if !ok || len(songs) == 0 {
		song_dict, _ = DBGetSongBySingerAndName(singer, song)
	} else {
		song_dict = songs[0]
	}

	if len(song_dict) > 0 {
		if song_dict["copyright"] == "1" {
			//resp.Hit = fmt.Sprintf("现在为您播放%s的%s", singer, song)
			data.HeaderContent = []string{"Name", "Singer", "PicUrl"}
			var data_content [][]string
			details := []string{song_dict["song"], song_dict["singer"], song_dict["picurl"]}
			data_content = append(data_content, details)
			data.DataContent = data_content

			resp.TextContent = data
			resp.Hidden = song_dict["url"]

			redis_key := fmt.Sprintf("%s:%s:song", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["song"])
			redis_key = fmt.Sprintf("%s:%s:singer", agentid, sessionid)
			RedisSetKeyValue(redis_key, song_dict["singer"])
		} else {
			resp.Service = "qa"
			resp.Operation = ""
			resp.Hit = "我还不会唱这首歌哦"
		}
	} else {
		// 后续可推荐该歌手的其他歌曲
		resp.Hit = fmt.Sprintf("抱歉,你上次听的什么歌来着?")
	}

	message_string, _ := json.Marshal(resp)
	return string(message_string)
}
