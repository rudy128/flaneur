package utils_twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Likes     int       `json:"likes"`
	Retweets  int       `json:"retweets"`
	Replies   int       `json:"replies"`
}

func ExtractTweetID(tweetURL string) string {
	re := regexp.MustCompile(`/status/(\d+)`)
	matches := re.FindStringSubmatch(tweetURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func GetAllTweetReplies(tweetID string) ([]*Tweet, error) {
	if !isLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	allTweets := []*Tweet{}
	cursor := ""
	maxPages := 200
	pageCount := 0

	for pageCount < maxPages {
		// Updated GraphQL operation ID (as of Nov 2025)
		req, _ := http.NewRequest("GET", "https://x.com/i/api/graphql/YVyS4SfwYW7Uw5qwy0mQCA/TweetDetail", nil)

		variables := map[string]interface{}{
			"focalTweetId":                           tweetID,
			"with_rux_injections":                    false,
			"rankingMode":                            "Relevance",
			"includePromotedContent":                 true,
			"withCommunity":                          true,
			"withQuickPromoteEligibilityTweetFields": true,
			"withBirdwatchNotes":                     true,
			"withVoice":                              true,
		}

		if cursor != "" {
			variables["cursor"] = cursor
			variables["count"] = 20
		}

		features := map[string]interface{}{
			"rweb_video_screen_enabled":                                               false,
			"payments_enabled":                                                        false,
			"profile_label_improvements_pcf_label_in_post_enabled":                    true,
			"responsive_web_profile_redirect_enabled":                                 false,
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

		fieldToggles := map[string]interface{}{
			"withArticleRichContentState": true,
			"withArticlePlainText":        false,
			"withGrokAnalyze":             false,
			"withDisallowedReplyControls": false,
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
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Twitter-Active-User", "yes")
		req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
		req.Header.Set("X-Twitter-Client-Language", "en")

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
		entryCount := countEntriesInResponse(result)

		fmt.Printf("\n=== PAGE %d RESPONSE ===\n", pageCount+1)
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Body Length: %d bytes\n", len(body))
		fmt.Printf("Total entries found: %d\n", entryCount)
		fmt.Printf("=== END PAGE %d RESPONSE ===\n\n", pageCount+1)

		if resp.StatusCode == 429 {
			fmt.Printf("Rate limited (429). Waiting 60 seconds...\n")
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			fmt.Printf("Error: Status %s\n", resp.Status)
			return allTweets, nil
		}

		pageTweets, nextCursor := parseTwitterResponse(result)
		allTweets = append(allTweets, pageTweets...)

		fmt.Printf("Page %d: Found %d tweets (Total: %d)\n", pageCount+1, len(pageTweets), len(allTweets))
		if len(pageTweets) > 0 {
			lastTweet := pageTweets[len(pageTweets)-1]
			fmt.Printf("Last tweet: @%s: %s\n", lastTweet.Username, lastTweet.Text[:min(50, len(lastTweet.Text))])
		}
		fmt.Printf("Current cursor: %s\n", cursor)
		fmt.Printf("Next cursor: %s\n", nextCursor)

		if nextCursor == "" || len(pageTweets) == 0 || nextCursor == cursor {
			fmt.Printf("No more pages available (cursor: %s, tweets: %d)\n", nextCursor, len(pageTweets))
			break
		}

		cursor = nextCursor
		pageCount++
		time.Sleep(2 * time.Second)
	}

	return allTweets, nil
}

func countEntriesInResponse(result map[string]interface{}) int {
	count := 0
	if data, ok := result["data"].(map[string]interface{}); ok {
		if threadedConversation, ok := data["threaded_conversation_with_injections_v2"].(map[string]interface{}); ok {
			if instructions, ok := threadedConversation["instructions"].([]interface{}); ok {
				for _, instruction := range instructions {
					if inst, ok := instruction.(map[string]interface{}); ok {
						if instType, ok := inst["type"].(string); ok && instType == "TimelineAddEntries" {
							if entries, ok := inst["entries"].([]interface{}); ok {
								for _, entry := range entries {
									if e, ok := entry.(map[string]interface{}); ok {
										if _, ok := e["entryId"].(string); ok {
											count++
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
	return count
}

func parseTwitterResponse(result map[string]interface{}) ([]*Tweet, string) {
	tweets := []*Tweet{}
	nextCursor := ""

	if data, ok := result["data"].(map[string]interface{}); ok {
		if threadedConversation, ok := data["threaded_conversation_with_injections_v2"].(map[string]interface{}); ok {
			if instructions, ok := threadedConversation["instructions"].([]interface{}); ok {
				for _, instruction := range instructions {
					if inst, ok := instruction.(map[string]interface{}); ok {
						if instType, ok := inst["type"].(string); ok && instType == "TimelineAddEntries" {
							if entries, ok := inst["entries"].([]interface{}); ok {
								for _, entry := range entries {
									if e, ok := entry.(map[string]interface{}); ok {
										if entryId, ok := e["entryId"].(string); ok {
											// Extract cursor for pagination
											if strings.Contains(entryId, "cursor-bottom") || strings.Contains(entryId, "cursor-showmorethreads") {
												if content, ok := e["content"].(map[string]interface{}); ok {
													if value, ok := content["value"].(string); ok {
														nextCursor = value
														fmt.Printf("Extracted cursor (%s): %s...\n", entryId, nextCursor[:50])
													}
												}
											}

											// Handle conversationthread entries (replies)
											if strings.Contains(entryId, "conversationthread") {
												if content, ok := e["content"].(map[string]interface{}); ok {
													if items, ok := content["items"].([]interface{}); ok {
														for _, item := range items {
															if itemObj, ok := item.(map[string]interface{}); ok {
																if itemContent, ok := itemObj["item"].(map[string]interface{}); ok {
																	if tweetContent, ok := itemContent["itemContent"].(map[string]interface{}); ok {
																		// Check if this is a showmore module and skip it
																		if typename, ok := tweetContent["__typename"].(string); ok {
																			if typename == "TimelineTimelineModule" || strings.Contains(typename, "ShowMore") {
																				continue
																			}
																		}
																		// Handle regular tweets
																		if tweetResults, ok := tweetContent["tweet_results"].(map[string]interface{}); ok {
																			if tweetResult, ok := tweetResults["result"].(map[string]interface{}); ok {
																				tweet := parseTweetFromResult(tweetResult)
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

											// Handle regular tweet entries
											if content, ok := e["content"].(map[string]interface{}); ok {
												if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
													if tweetResults, ok := itemContent["tweet_results"].(map[string]interface{}); ok {
														if tweetResult, ok := tweetResults["result"].(map[string]interface{}); ok {
															tweet := parseTweetFromResult(tweetResult)
															if tweet != nil && !strings.Contains(entryId, "tweet-1967514277054738899") {
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

	return tweets, nextCursor
}

func parseTweetFromResult(result map[string]interface{}) *Tweet {
	if legacy, ok := result["legacy"].(map[string]interface{}); ok {
		if core, ok := result["core"].(map[string]interface{}); ok {
			if userResults, ok := core["user_results"].(map[string]interface{}); ok {
				if userResult, ok := userResults["result"].(map[string]interface{}); ok {
					if userLegacy, ok := userResult["legacy"].(map[string]interface{}); ok {
						tweet := &Tweet{}

						if id, ok := legacy["id_str"].(string); ok {
							tweet.ID = id
						}
						if text, ok := legacy["full_text"].(string); ok {
							tweet.Text = text
						}
						if username, ok := userLegacy["screen_name"].(string); ok {
							tweet.Username = username
						}
						if name, ok := userLegacy["name"].(string); ok {
							tweet.Name = name
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

						return tweet
					}
				}
			}
		}
	}
	return nil
}
