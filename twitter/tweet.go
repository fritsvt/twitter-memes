package twitter

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/ChimeraCoder/anaconda"
	"github.com/devedge/imagehash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"twitter-meme-bot/database"
	"twitter-meme-bot/structs"
)

var (
	api anaconda.TwitterApi
	hashTag = ""
)

func Setup() {
	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_PUBLIC"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))
	api = *anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

	hashTag = os.Getenv("HASH_TAG")
}

func SendTweet(thread structs.Thread, checkImageHash bool, scheduled *structs.ScheduledTweet) {
	response, e := http.Get(thread.ImageUrl)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		str := base64.StdEncoding.EncodeToString([]byte(bodyString))

		hash := ""
		if checkImageHash {
			hash = getImageHash([]byte(bodyString), thread)
			thread.ImageHash = hash
		}

		if !checkImageHash || !database.GetThreadByHash(hash) {
			if checkImageHash { database.InsertThread(thread) }

			res, err := api.UploadMedia(str);
			if err != nil {
				log.Fatal(err)
			}

			v := url.Values{}
			if scheduled.ID != 0 {
				v.Set("in_reply_to_status_id", scheduled.StatusId)
				thread.Title = "@" + scheduled.ToUser + " " + thread.Title
			} else {
				if hashTag != "" {
					thread.Title = thread.Title + " #" + hashTag
					println(thread.Title)
				}
			}
			v.Set("media_ids", strconv.FormatInt(res.MediaID, 10))

			println("Posting tweet: " + thread.RedditId)
			api.PostTweet(thread.Title,  v);
		} else { 
			println("duplicate image found: " + thread.RedditId)
		}
	}

}

func getImageHash(image []byte, thread structs.Thread) (hash string) {
	err := ioutil.WriteFile("./tmp/"+thread.RedditId+thread.Extension, image, 0644)
	if err != nil {
		log.Fatal(err)
	}

	src, _ := imagehash.OpenImg("./tmp/"+thread.RedditId+thread.Extension)
	rawHash, _ := imagehash.Dhash(src, 16)

	os.Remove("./tmp/"+thread.RedditId+thread.Extension)

	return hex.EncodeToString(rawHash)
}