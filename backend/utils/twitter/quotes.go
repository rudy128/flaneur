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

func SearchQuotedTweets(quotedTweetID string) ([]*Tweet, error) {
	if !isLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	allTweets := []*Tweet{}
	cursor := ""
	maxPages := 50
	pageCount := 0

	for pageCount < maxPages {
		req, _ := http.NewRequest("GET", "https://x.com/i/api/graphql/7fWgap3nJOk9UpFV7UqcoQ/SearchTimeline", nil)

		variables := map[string]interface{}{
			"rawQuery":              fmt.Sprintf("quoted_tweet_id:%s", quotedTweetID),
			"count":                 20,
			"querySource":           "tdqt",
			"product":               "Latest",
			"withGrokTranslatedBio": false,
		}

		if cursor != "" {
			variables["cursor"] = cursor
		}

		features := map[string]interface{}{
			"rweb_video_screen_enabled":                                               false,
			"payments_enabled":                                                        false,
			"profile_label_improvements_pcf_label_in_post_enabled":                    true,
			"rweb_tipjar_consumption_enabled":                                         true,
			"verified_phone_label_enabled":                                            true,
			"creator_subscriptions_tweet_preview_api_enabled":                         true,
			"responsive_web_graphql_timeline_navigation_enabled":                      true,
			"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
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
			"responsive_web_grok_image_annotation_enabled":                            true,
			"responsive_web_grok_imagine_annotation_enabled":                          true,
			"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
			"responsive_web_enhance_cards_enabled":                                    false,
		}

		query := url.Values{}
		variablesJSON, _ := json.Marshal(variables)
		featuresJSON, _ := json.Marshal(features)

		query.Set("variables", string(variablesJSON))
		query.Set("features", string(featuresJSON))
		req.URL.RawQuery = query.Encode()

		req.Header.Set("Authorization", "Bearer "+bearerToken2)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("X-Guest-Token", globalGuestToken)
		req.Header.Set("X-Twitter-Active-User", "yes")
		req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
		req.Header.Set("X-Twitter-Client-Language", "en")
		req.Header.Set("Referer", "https://x.com/CubaneSpace/status/1955622870077309307/quotes")

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

		fmt.Printf("\n=== PAGE %d RESPONSE ===\n", pageCount+1)
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Body Length: %d bytes\n", len(body))
		fmt.Printf("=== END PAGE %d RESPONSE ===\n\n", pageCount+1)

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if resp.StatusCode == 429 {
			fmt.Printf("Rate limited (429). Waiting 60 seconds...\n")
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			fmt.Printf("Error: Status %s\n", resp.Status)
			return allTweets, nil
		}

		pageTweets, nextCursor := parseSearchResponse(result)
		allTweets = append(allTweets, pageTweets...)

		fmt.Printf("Page %d: Found %d tweets (Total: %d)\n", pageCount+1, len(pageTweets), len(allTweets))
		if len(pageTweets) > 0 {
			lastTweet := pageTweets[len(pageTweets)-1]
			fmt.Printf("Last tweet: @%s: %s\n", lastTweet.Username, lastTweet.Text[:min(50, len(lastTweet.Text))])
		}
		fmt.Printf("Current cursor: %s\n", cursor)
		fmt.Printf("Next cursor: %s\n", nextCursor)
		fmt.Printf("Cursor same as previous: %t\n", nextCursor == cursor)
		fmt.Printf("Next cursor empty: %t\n", nextCursor == "")
		fmt.Printf("No tweets found: %t\n", len(pageTweets) == 0)

		if nextCursor == "" {
			fmt.Printf("DEBUG: No cursor found in response\n")
		}

		if nextCursor == "" || len(pageTweets) == 0 || nextCursor == cursor {
			fmt.Printf("Stopping pagination - Reason: ")
			if nextCursor == "" {
				fmt.Printf("No next cursor\n")
			} else if len(pageTweets) == 0 {
				fmt.Printf("No tweets found\n")
			} else if nextCursor == cursor {
				fmt.Printf("Cursor unchanged\n")
			}
			break
		}

		cursor = nextCursor
		pageCount++
		time.Sleep(2 * time.Second)
	}

	return allTweets, nil
}

func parseSearchResponse(result map[string]interface{}) ([]*Tweet, string) {
	tweets := []*Tweet{}
	nextCursor := ""

	if data, ok := result["data"].(map[string]interface{}); ok {
		if search, ok := data["search_by_raw_query"].(map[string]interface{}); ok {
			if timeline, ok := search["search_timeline"].(map[string]interface{}); ok {
				if timeline_obj, ok := timeline["timeline"].(map[string]interface{}); ok {
					if instructions, ok := timeline_obj["instructions"].([]interface{}); ok {
						for _, instruction := range instructions {
							if inst, ok := instruction.(map[string]interface{}); ok {
								if instType, ok := inst["type"].(string); ok && (instType == "TimelineAddEntries" || instType == "TimelineReplaceEntry") {
									if instType == "TimelineAddEntries" {
										if entries, ok := inst["entries"].([]interface{}); ok {
											for _, entry := range entries {
												if e, ok := entry.(map[string]interface{}); ok {
													if entryId, ok := e["entryId"].(string); ok {
														if strings.Contains(entryId, "cursor-bottom") {
															if content, ok := e["content"].(map[string]interface{}); ok {
																if cursorType, ok := content["cursorType"].(string); ok {
																	if value, ok := content["value"].(string); ok {
																		if cursorType == "Bottom" {
																			nextCursor = value
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
									if instType == "TimelineReplaceEntry" {
										if entryIdToReplace, ok := inst["entry_id_to_replace"].(string); ok {
											if strings.Contains(entryIdToReplace, "cursor-bottom") {
												if entry, ok := inst["entry"].(map[string]interface{}); ok {
													if content, ok := entry["content"].(map[string]interface{}); ok {
														if value, ok := content["value"].(string); ok {
															nextCursor = value
															fmt.Printf("DEBUG: Found cursor in TimelineReplaceEntry for bottom: %s\n", value)
														}
													}
												}
											}
										}
									}
									if instType == "TimelineAddEntries" {
										if entries, ok := inst["entries"].([]interface{}); ok {
											for _, entry := range entries {
												if e, ok := entry.(map[string]interface{}); ok {
													if entryId, ok := e["entryId"].(string); ok {
														if strings.Contains(entryId, "tweet-") {
															if content, ok := e["content"].(map[string]interface{}); ok {
																if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
																	if tweetResults, ok := itemContent["tweet_results"].(map[string]interface{}); ok {
																		if tweetResult, ok := tweetResults["result"].(map[string]interface{}); ok {
																			tweet := parseSearchTweetResult(tweetResult)
																			if tweet != nil {
																				tweets = append(tweets, tweet)
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
							}
						}
					}
				}
			}
		}
	}

	return tweets, nextCursor
}

func parseSearchTweetResult(result map[string]interface{}) *Tweet {
	tweet := &Tweet{}

	if legacy, ok := result["legacy"].(map[string]interface{}); ok {
		if id, ok := legacy["id_str"].(string); ok {
			tweet.ID = id
		}
		if text, ok := legacy["full_text"].(string); ok {
			tweet.Text = text
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
	}

	if core, ok := result["core"].(map[string]interface{}); ok {
		if userResults, ok := core["user_results"].(map[string]interface{}); ok {
			if userResult, ok := userResults["result"].(map[string]interface{}); ok {
				if userCore, ok := userResult["core"].(map[string]interface{}); ok {
					if username, ok := userCore["screen_name"].(string); ok {
						tweet.Username = username
					}
					if name, ok := userCore["name"].(string); ok {
						tweet.Name = name
					}
				}
				if tweet.Username == "" || tweet.Name == "" {
					if userLegacy, ok := userResult["legacy"].(map[string]interface{}); ok {
						if tweet.Username == "" {
							if username, ok := userLegacy["screen_name"].(string); ok {
								tweet.Username = username
							}
						}
						if tweet.Name == "" {
							if name, ok := userLegacy["name"].(string); ok {
								tweet.Name = name
							}
						}
					}
				}
			}
		}
	}

	if tweet.ID != "" {
		return tweet
	}
	return nil
}
