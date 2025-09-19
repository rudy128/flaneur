package utils_twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TweetData represents the extracted tweet information
type TweetData struct {
	Content string     `json:"content"`
	Media   MediaData  `json:"media"`
	Author  AuthorData `json:"author"`
}

// MediaData represents media content in a tweet
type MediaData struct {
	Photos []string `json:"photos"`
	Videos []string `json:"videos"`
	GIFs   []string `json:"gifs"`
}

// AuthorData represents tweet author information
type AuthorData struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

// ExtractTweetContent extracts the text content from a tweet
func ExtractTweetContent(tweet *Tweet) string {
	if tweet == nil {
		return ""
	}
	return strings.TrimSpace(tweet.Text)
}

// ExtractTweetMedia extracts all media URLs from a TweetWithMedia
func ExtractTweetMediaFromFull(tweetWithMedia *TweetWithMedia) MediaData {
	media := MediaData{
		Photos: make([]string, 0),
		Videos: make([]string, 0),
		GIFs:   make([]string, 0),
	}

	if tweetWithMedia == nil {
		return media
	}

	media.Photos = tweetWithMedia.Photos
	media.Videos = tweetWithMedia.Videos
	media.GIFs = tweetWithMedia.GIFs

	return media
}

// ExtractTweetMedia extracts all media URLs from a tweet (legacy)
func ExtractTweetMedia(tweet *Tweet) MediaData {
	media := MediaData{
		Photos: make([]string, 0),
		Videos: make([]string, 0),
		GIFs:   make([]string, 0),
	}

	if tweet == nil {
		return media
	}

	// Media extraction would need to be implemented based on tweet parsing
	// Current Tweet struct doesn't include media fields
	return media
}

// ExtractTweetAuthorFromFull extracts author information from TweetWithMedia
func ExtractTweetAuthorFromFull(tweetWithMedia *TweetWithMedia) AuthorData {
	author := AuthorData{}

	if tweetWithMedia == nil || tweetWithMedia.Tweet == nil {
		return author
	}

	author.Username = tweetWithMedia.Tweet.Username
	author.Name = tweetWithMedia.Tweet.Name
	author.Avatar = tweetWithMedia.Avatar

	return author
}

// ExtractTweetAuthor extracts author information from a tweet (legacy)
func ExtractTweetAuthor(tweet *Tweet) AuthorData {
	author := AuthorData{}

	if tweet == nil {
		return author
	}

	author.Username = tweet.Username
	author.Name = tweet.Name
	// Avatar would need to be extracted from the full API response
	author.Avatar = ""

	return author
}

// ExtractFullTweetData extracts all tweet data in one function
func ExtractFullTweetData(tweet *Tweet) TweetData {
	return TweetData{
		Content: ExtractTweetContent(tweet),
		Media:   ExtractTweetMedia(tweet),
		Author:  ExtractTweetAuthor(tweet),
	}
}

// GetTweetAge calculates how long ago the tweet was posted
func GetTweetAge(tweet *Tweet) time.Duration {
	if tweet == nil {
		return 0
	}
	return time.Since(tweet.Timestamp)
}

// HasMedia checks if tweet contains any media
func HasMedia(tweet *Tweet) bool {
	if tweet == nil {
		return false
	}
	// Media detection would need to be implemented based on tweet text parsing
	return false
}

// IsReply checks if tweet is a reply
func IsReply(tweet *Tweet) bool {
	if tweet == nil {
		return false
	}
	// Reply detection would need to be implemented based on tweet text parsing
	return strings.HasPrefix(tweet.Text, "@")
}

// IsRetweet checks if tweet is a retweet
func IsRetweet(tweet *Tweet) bool {
	if tweet == nil {
		return false
	}
	// Retweet detection would need to be implemented based on tweet text parsing
	return strings.HasPrefix(tweet.Text, "RT @")
}

// GetEngagementStats returns likes, retweets, and replies count
func GetEngagementStats(tweet *Tweet) (int, int, int) {
	if tweet == nil {
		return 0, 0, 0
	}
	return tweet.Likes, tweet.Retweets, tweet.Replies
}

// GetTweetData fetches and extracts complete tweet data by ID
func GetTweetData(tweetID string) (*TweetData, error) {
	if !isLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	tweet, err := fetchSingleTweet(tweetID)
	if err != nil {
		return nil, err
	}

	if tweet == nil {
		return nil, fmt.Errorf("tweet not found")
	}

	tweetData := ExtractFullTweetData(tweet)
	return &tweetData, nil
}

// fetchSingleTweet fetches a single tweet by ID using Twitter API
func fetchSingleTweet(tweetID string) (*Tweet, error) {
	req, err := http.NewRequest("GET", "https://x.com/i/api/graphql/wqi5M7wZ7tW-X9S2t-Mqcg/TweetResultByRestId", nil)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"tweetId":                tweetID,
		"includePromotedContent": true,
		"withBirdwatchNotes":     true,
		"withVoice":              true,
		"withCommunity":          true,
	}

	features := map[string]interface{}{
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"premium_content_api_read_enabled":                                        false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"responsive_web_grok_analyze_button_fetch_trends_enabled":                 false,
		"responsive_web_grok_analyze_post_followups_enabled":                      true,
		"responsive_web_jetfuel_frame":                                            true,
		"responsive_web_grok_share_attachment_enabled":                            true,
		"articles_preview_enabled":                                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"responsive_web_grok_show_grok_translated_post":                           false,
		"responsive_web_grok_analysis_button_from_backend":                        true,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"payments_enabled":                                                        false,
		"profile_label_improvements_pcf_label_in_post_enabled":                    true,
		"rweb_tipjar_consumption_enabled":                                         true,
		"verified_phone_label_enabled":                                            true,
		"responsive_web_grok_image_annotation_enabled":                            true,
		"responsive_web_grok_imagine_annotation_enabled":                          true,
		"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	fieldToggles := map[string]interface{}{
		"withArticleRichContentState": true,
		"withArticlePlainText":        false,
	}

	query := url.Values{}
	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(features)
	fieldTogglesJSON, _ := json.Marshal(fieldToggles)

	query.Set("variables", string(variablesJSON))
	query.Set("features", string(featuresJSON))
	query.Set("fieldToggles", string(fieldTogglesJSON))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", "Bearer "+bearerToken2)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Guest-Token", globalGuestToken)
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
	req.Header.Set("X-Twitter-Client-Language", "en")
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range globalClient.Jar.Cookies(req.URL) {
		if cookie.Name == "ct0" {
			req.Header.Set("X-CSRF-Token", cookie.Value)
			break
		}
	}

	resp, err := globalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	// Parse the tweet from the response
	tweet := parseTweetResultByRestId(result)
	return tweet, nil
}

// parseTweetResultByRestId parses a tweet from TweetResultByRestId API response
func parseTweetResultByRestId(result map[string]interface{}) *Tweet {
	if data, ok := result["data"].(map[string]interface{}); ok {
		if tweetResult, ok := data["tweetResult"].(map[string]interface{}); ok {
			if tweetData, ok := tweetResult["result"].(map[string]interface{}); ok {
				return parseTweetFromTweetResult(tweetData)
			}
		}
	}
	return nil
}

// TweetWithMedia extends Tweet with media information
type TweetWithMedia struct {
	*Tweet
	Photos []string `json:"photos"`
	Videos []string `json:"videos"`
	GIFs   []string `json:"gifs"`
	Avatar string   `json:"avatar"`
}

// parseTweetFromTweetResult parses tweet data from the result object
func parseTweetFromTweetResult(result map[string]interface{}) *Tweet {
	tweet := &Tweet{}

	// Get basic tweet info
	if restId, ok := result["rest_id"].(string); ok {
		tweet.ID = restId
	}

	// Get legacy data
	if legacy, ok := result["legacy"].(map[string]interface{}); ok {
		if fullText, ok := legacy["full_text"].(string); ok {
			tweet.Text = fullText
		}
		if likes, ok := legacy["favorite_count"].(float64); ok {
			tweet.Likes = int(likes)
		}
		if retweets, ok := legacy["retweet_count"].(float64); ok {
			tweet.Retweets = int(retweets)
		}
		if replies, ok := legacy["reply_count"].(float64); ok {
			tweet.Replies = int(replies)
		}
		if createdAt, ok := legacy["created_at"].(string); ok {
			if parsedTime, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", createdAt); err == nil {
				tweet.Timestamp = parsedTime
			}
		}
	}

	// Get user info from core
	if core, ok := result["core"].(map[string]interface{}); ok {
		if userResults, ok := core["user_results"].(map[string]interface{}); ok {
			if userResult, ok := userResults["result"].(map[string]interface{}); ok {
				if userCore, ok := userResult["core"].(map[string]interface{}); ok {
					if name, ok := userCore["name"].(string); ok {
						tweet.Name = name
					}
					if screenName, ok := userCore["screen_name"].(string); ok {
						tweet.Username = screenName
					}
				}
			}
		}
	}

	return tweet
}

// GetTweetWithMedia fetches tweet with media and avatar information
func GetTweetWithMedia(tweetID string) (*TweetWithMedia, error) {
	if !isLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	req, err := http.NewRequest("GET", "https://x.com/i/api/graphql/wqi5M7wZ7tW-X9S2t-Mqcg/TweetResultByRestId", nil)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"tweetId":                tweetID,
		"includePromotedContent": true,
		"withBirdwatchNotes":     true,
		"withVoice":              true,
		"withCommunity":          true,
	}

	features := map[string]interface{}{
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"premium_content_api_read_enabled":                                        false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"responsive_web_grok_analyze_button_fetch_trends_enabled":                 false,
		"responsive_web_grok_analyze_post_followups_enabled":                      true,
		"responsive_web_jetfuel_frame":                                            true,
		"responsive_web_grok_share_attachment_enabled":                            true,
		"articles_preview_enabled":                                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"responsive_web_grok_show_grok_translated_post":                           false,
		"responsive_web_grok_analysis_button_from_backend":                        true,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"payments_enabled":                                                        false,
		"profile_label_improvements_pcf_label_in_post_enabled":                    true,
		"rweb_tipjar_consumption_enabled":                                         true,
		"verified_phone_label_enabled":                                            true,
		"responsive_web_grok_image_annotation_enabled":                            true,
		"responsive_web_grok_imagine_annotation_enabled":                          true,
		"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	fieldToggles := map[string]interface{}{
		"withArticleRichContentState": true,
		"withArticlePlainText":        false,
	}

	query := url.Values{}
	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(features)
	fieldTogglesJSON, _ := json.Marshal(fieldToggles)

	query.Set("variables", string(variablesJSON))
	query.Set("features", string(featuresJSON))
	query.Set("fieldToggles", string(fieldTogglesJSON))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", "Bearer "+bearerToken2)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Guest-Token", globalGuestToken)
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
	req.Header.Set("X-Twitter-Client-Language", "en")
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range globalClient.Jar.Cookies(req.URL) {
		if cookie.Name == "ct0" {
			req.Header.Set("X-CSRF-Token", cookie.Value)
			break
		}
	}

	resp, err := globalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	// Parse the full tweet with media
	tweetWithMedia := parseFullTweetResult(result)
	return tweetWithMedia, nil
}

// parseFullTweetResult parses complete tweet data including media and avatar
func parseFullTweetResult(result map[string]interface{}) *TweetWithMedia {
	tweetWithMedia := &TweetWithMedia{
		Tweet:  &Tweet{},
		Photos: make([]string, 0),
		Videos: make([]string, 0),
		GIFs:   make([]string, 0),
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if tweetResult, ok := data["tweetResult"].(map[string]interface{}); ok {
			if tweetData, ok := tweetResult["result"].(map[string]interface{}); ok {
				// Parse basic tweet info
				if restId, ok := tweetData["rest_id"].(string); ok {
					tweetWithMedia.Tweet.ID = restId
				}

				// Parse legacy data
				if legacy, ok := tweetData["legacy"].(map[string]interface{}); ok {
					if fullText, ok := legacy["full_text"].(string); ok {
						tweetWithMedia.Tweet.Text = fullText
					}
					if likes, ok := legacy["favorite_count"].(float64); ok {
						tweetWithMedia.Tweet.Likes = int(likes)
					}
					if retweets, ok := legacy["retweet_count"].(float64); ok {
						tweetWithMedia.Tweet.Retweets = int(retweets)
					}
					if replies, ok := legacy["reply_count"].(float64); ok {
						tweetWithMedia.Tweet.Replies = int(replies)
					}
					if createdAt, ok := legacy["created_at"].(string); ok {
						if parsedTime, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", createdAt); err == nil {
							tweetWithMedia.Tweet.Timestamp = parsedTime
						}
					}

					// Parse media from extended_entities
					if extendedEntities, ok := legacy["extended_entities"].(map[string]interface{}); ok {
						if media, ok := extendedEntities["media"].([]interface{}); ok {
							for _, mediaItem := range media {
								if mediaObj, ok := mediaItem.(map[string]interface{}); ok {
									if mediaType, ok := mediaObj["type"].(string); ok {
										switch mediaType {
										case "photo":
											if mediaURL, ok := mediaObj["media_url_https"].(string); ok {
												tweetWithMedia.Photos = append(tweetWithMedia.Photos, mediaURL)
											}
										case "video":
											if videoInfo, ok := mediaObj["video_info"].(map[string]interface{}); ok {
												if variants, ok := videoInfo["variants"].([]interface{}); ok {
													for _, variant := range variants {
														if variantObj, ok := variant.(map[string]interface{}); ok {
															if contentType, ok := variantObj["content_type"].(string); ok && contentType == "video/mp4" {
																if videoURL, ok := variantObj["url"].(string); ok {
																	tweetWithMedia.Videos = append(tweetWithMedia.Videos, videoURL)
																	break
																}
															}
														}
													}
												}
											}
										case "animated_gif":
											if videoInfo, ok := mediaObj["video_info"].(map[string]interface{}); ok {
												if variants, ok := videoInfo["variants"].([]interface{}); ok {
													for _, variant := range variants {
														if variantObj, ok := variant.(map[string]interface{}); ok {
															if gifURL, ok := variantObj["url"].(string); ok {
																tweetWithMedia.GIFs = append(tweetWithMedia.GIFs, gifURL)
																break
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}

				// Parse user info and avatar
				if core, ok := tweetData["core"].(map[string]interface{}); ok {
					if userResults, ok := core["user_results"].(map[string]interface{}); ok {
						if userResult, ok := userResults["result"].(map[string]interface{}); ok {
							if userCore, ok := userResult["core"].(map[string]interface{}); ok {
								if name, ok := userCore["name"].(string); ok {
									tweetWithMedia.Tweet.Name = name
								}
								if screenName, ok := userCore["screen_name"].(string); ok {
									tweetWithMedia.Tweet.Username = screenName
								}
							}
							// Get avatar from user result
							if avatar, ok := userResult["avatar"].(map[string]interface{}); ok {
								if imageURL, ok := avatar["image_url"].(string); ok {
									tweetWithMedia.Avatar = imageURL
								}
							}
						}
					}
				}
			}
		}
	}

	return tweetWithMedia
}
