package gosubtitles

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// GetSubtitles extraxt Youtube subtitles
// Youtube Subtitle extraxtion
// param: video URL string
// param: lang	 string
func GetSubtitles(videoURL, lang string, timeout int) (string, error, int) {
	var caption map[string]interface{}
	if videoURL == "" {
		return "", fmt.Errorf("you have mistake on your url address"), 0
	}
	if lang == "" {
		lang = "ko"
	}
	// Extraxt video id from URL string
	regexVideoID, _ := regexp.Compile(`(?:https?:\/{2})?(?:w{3}\.)?youtu(?:be)?\.(?:com|be)(?:\/watch\?v=|\/)([^\s&]+)`)
	videoID := regexVideoID.FindStringSubmatch(videoURL)
	// Get video info
	time.Sleep(time.Second * time.Duration(timeout))
	data, err := http.Get(fmt.Sprintf("https://www.youtube.com/get_video_info?video_id=%s", videoID[1]))
	if nil != err {
		fmt.Println(fmt.Sprintf("Error occured while getting youtube caption address: %s", err))
		return "", err, 0
	}
	defer data.Body.Close()
	body, err := ioutil.ReadAll(data.Body)
	if err != nil {
		fmt.Println(err)
	}
	if "" == string(body) {
		return "", fmt.Errorf("Too many Requests"), 0
	}
	// decode response
	decodedData, err := url.QueryUnescape(string(body))
	if nil != err {
		fmt.Println(fmt.Sprintf("Error while decoding string: %s", err))
		return "", err, 0
	}
	// check if captions exists
	if strings.Contains(decodedData, "captionTracks") == false {
		//	t.Log(t.Warning, fmt.Sprintf("could not find captions for video by id %s", videoURL))
		return "", fmt.Errorf("could not find captions for video"), 0
	}
	// extract caption json string
	regex, _ := regexp.Compile(`({"captionTracks":.*isTranslatable":(true|false)}])`)

	strMatch := regex.FindString(decodedData)
	strMatch += "}"
	if err = json.Unmarshal([]byte(strMatch), &caption); nil != err {
		fmt.Println(fmt.Sprintf("Error occured while parsing json: %s", err))
		return "", err, 0
	}
	baseCaption := caption["captionTracks"].([]interface{})

	for key, result := range baseCaption {
		baseCaptionDetails := baseCaption[key].(map[string]interface{})
		if baseCaptionDetails["vssId"] == "a."+lang || baseCaptionDetails["vssId"] == lang {
			captionData, err := http.Get(baseCaptionDetails["baseUrl"].(string))
			if nil != err || result == nil {
				fmt.Println(fmt.Sprintf("Error getting captions: %s", err))
				return "", err, 0
			}
			defer captionData.Body.Close()
			captionBody, err := ioutil.ReadAll(captionData.Body)
			if err != nil {
				fmt.Println(err)
			}
			captBody := string(captionBody)
			// remove <xml> symbols from caption
			textRegex, _ := regexp.Compile(`<text start="([\d.]+)" dur="([\d.]+)">`)
			captBody = strings.ReplaceAll(captBody, `<?xml version="1.0" encoding="utf-8" ?><transcript>`, "")
			captBody = textRegex.ReplaceAllString(captBody, " ")
			captBody = strings.ReplaceAll(captBody, "</text>", "")
			captBody = strings.ReplaceAll(captBody, "</transcript>", "")
			captBody = html.UnescapeString(captBody)
			// response raw text of captions
			return captBody, err, 1
		}
	}
	return "", err, 0
}
