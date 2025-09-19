package utils_twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func getLikers(tweetID string) error {
	if !isLoggedIn {
		return fmt.Errorf("not logged in")
	}

	allUsers := []*User{}
	cursor := ""
	maxPages := 50
	pageCount := 0

	for pageCount < maxPages {
		req, _ := http.NewRequest("GET", "https://x.com/i/api/graphql/O_NkyaaOXLOMaPFG1BKJNA/Favoriters", nil)

		variables := map[string]interface{}{
			"tweetId":                tweetID,
			"count":                  20,
			"includePromotedContent": true,
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
		req.Header.Set("Referer", fmt.Sprintf("https://x.com/Kiyotaka1232384/status/%s/likes", tweetID))

		for _, cookie := range globalClient.Jar.Cookies(req.URL) {
			if cookie.Name == "ct0" {
				req.Header.Set("X-CSRF-Token", cookie.Value)
				break
			}
		}

		resp, err := globalClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		fmt.Printf("\n=== PAGE %d RESPONSE ===\n", pageCount+1)
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Body Length: %d bytes\n", len(body))
		fmt.Printf("=== END PAGE %d RESPONSE ===\n\n", pageCount+1)

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if resp.StatusCode != 200 {
			fmt.Printf("Error: Status %s\n", resp.Status)
			return fmt.Errorf("API error: %s", resp.Status)
		}

		pageUsers, nextCursor := parseLikersResponse(result)
		allUsers = append(allUsers, pageUsers...)

		fmt.Printf("Page %d: Found %d users (Total: %d)\n", pageCount+1, len(pageUsers), len(allUsers))
		fmt.Printf("Current cursor: %s\n", cursor)
		fmt.Printf("Next cursor: %s\n", nextCursor)
		fmt.Printf("Cursor same as previous: %t\n", nextCursor == cursor)
		fmt.Printf("Next cursor empty: %t\n", nextCursor == "")
		fmt.Printf("No users found: %t\n", len(pageUsers) == 0)

		if nextCursor == "" || len(pageUsers) == 0 || nextCursor == cursor {
			fmt.Printf("Stopping pagination - Reason: ")
			if nextCursor == "" {
				fmt.Printf("No next cursor\n")
			} else if len(pageUsers) == 0 {
				fmt.Printf("No users found\n")
			} else if nextCursor == cursor {
				fmt.Printf("Cursor unchanged\n")
			}
			break
		}

		cursor = nextCursor
		pageCount++
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\n=== FOUND %d TOTAL LIKERS ===\n", len(allUsers))

	usersJSON, err := json.MarshalIndent(allUsers, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling users to JSON: %v", err)
	}

	err = os.WriteFile("likes.json", usersJSON, 0644)
	if err != nil {
		return fmt.Errorf("error writing users to file: %v", err)
	}

	fmt.Printf("Likers saved to likes.json\n")
	return nil
}

func parseLikersResponse(result map[string]interface{}) ([]*User, string) {
	users := []*User{}
	nextCursor := ""

	if data, ok := result["data"].(map[string]interface{}); ok {
		if favoritersTimeline, ok := data["favoriters_timeline"].(map[string]interface{}); ok {
			if timeline, ok := favoritersTimeline["timeline"].(map[string]interface{}); ok {
				if instructions, ok := timeline["instructions"].([]interface{}); ok {
					for _, instruction := range instructions {
						if inst, ok := instruction.(map[string]interface{}); ok {
							if instType, ok := inst["type"].(string); ok && instType == "TimelineAddEntries" {
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
												} else if strings.Contains(entryId, "user-") {
													if content, ok := e["content"].(map[string]interface{}); ok {
														if itemContent, ok := content["itemContent"].(map[string]interface{}); ok {
															if userResults, ok := itemContent["user_results"].(map[string]interface{}); ok {
																if userResult, ok := userResults["result"].(map[string]interface{}); ok {
																	user := parseUserResult(userResult)
																	if user != nil {
																		users = append(users, user)
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

	return users, nextCursor
}
