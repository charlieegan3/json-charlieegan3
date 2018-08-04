package collect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

type profilePage struct {
	EntryData struct {
		ProfilePage []struct {
			Graphql struct {
				User struct {
					EdgeOwnerToTimelineMedia struct {
						Edges []struct {
							Node struct {
								Shortcode string `json:"shortcode"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_owner_to_timeline_media"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
}

type postPage struct {
	EntryData struct {
		PostPage []struct {
			Graphql struct {
				ShortcodeMedia struct {
					IsVideo          bool `json:"is_video"`
					TakenAtTimestamp int64
					Location         struct {
						Name string `json:"name"`
					} `json:"location"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
}

// LatestPo stores the URL, location and time of the latest post, video or photo
type LatestPost struct {
	URL       string    `json:"url"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"`
}

// Instagram returns latest post for a given user
func Instagram(host string, username string) (LatestPost, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", host, username))

	if err != nil {
		return LatestPost{}, errors.Wrap(err, "get profile page failed")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return LatestPost{}, errors.Wrap(err, "profile page body read failed")
	}

	var profilePageData profilePage
	err = parsePageJSON(body, &profilePageData)
	if err != nil {
		return LatestPost{}, errors.Wrap(err, "profile page json parsing failed")
	}
	fmt.Printf("%+v\n", profilePageData)
	if len(profilePageData.EntryData.ProfilePage) == 0 || len(profilePageData.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges) == 0 {
		return LatestPost{}, errors.New("profile page json invalid")
	}
	shortcode := profilePageData.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges[0].Node.Shortcode

	postURL := fmt.Sprintf("%s/p/%s", host, shortcode)
	resp, err = http.Get(postURL)
	if err != nil {
		return LatestPost{}, errors.Wrap(err, "get post page failed")
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return LatestPost{}, errors.Wrap(err, "post page body read failed")
	}

	var postPageData postPage
	err = parsePageJSON(body, &postPageData)
	if err != nil {
		return LatestPost{}, errors.Wrap(err, "post page json parsing failed")
	}
	if len(postPageData.EntryData.PostPage) == 0 {
		return LatestPost{}, errors.New("post page json invalid")
	}

	post := postPageData.EntryData.PostPage[0].Graphql.ShortcodeMedia
	postType := "photo"
	if post.IsVideo == true {
		postType = "video"
	}
	createdAt := time.Unix(post.TakenAtTimestamp, 0)

	return LatestPost{
		Location:  post.Location.Name,
		Type:      postType,
		URL:       postURL,
		CreatedAt: createdAt,
	}, nil
}

func parsePageJSON(body []byte, data interface{}) error {
	r := regexp.MustCompile("window._sharedData = (?P<Data>.*);</script>")
	matches := r.FindSubmatch(body)
	if len(matches) != 2 {
		return errors.New("unable to extract shared data from dom")
	}
	fmt.Println(string(matches[1]))

	err := json.Unmarshal(matches[1], data)
	if err != nil {
		return errors.Wrap(err, "shared data unmarshal failed")
	}

	return nil
}