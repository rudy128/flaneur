package utils_twitter

import (
	"fmt"
)

func main() {
	userID := "test-user-id" // This would normally come from your application context
	err := LoadTokensForUser(userID)
	needLogin := false

	if err != nil {
		fmt.Printf("No saved session found (%v)\n", err)
		needLogin = true
	} else {
		if !ValidateSession() {
			fmt.Println("Saved session is invalid/expired")
			needLogin = true
		}
	}

	if needLogin {
		fmt.Println("⚠️  WARNING: Will perform username/password login which may trigger account restrictions!")

		username := "AnelloJosh"
		password := "joshanello21@gmail..com"

		err = LoginAndSaveTokens(username, password, userID)
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}
	}

	tweetURL := "https://x.com/naiivememe/status/1968813787471077498"
	tweetID := ExtractTweetID(tweetURL)
	if tweetID == "" {
		fmt.Printf("Could not extract tweet ID from URL: %s\n", tweetURL)
		return
	}

	tweetWithMedia, err := GetTweetWithMedia(tweetID)
	if err != nil {
		fmt.Printf("Error getting tweet data: %v\n", err)
		return
	}

	fmt.Printf("Tweet content: %s\n", tweetWithMedia.Tweet.Text)
	fmt.Printf("Tweet author: %s (@%s)\n", tweetWithMedia.Tweet.Name, tweetWithMedia.Tweet.Username)
	fmt.Printf("Author avatar: %s\n", tweetWithMedia.Avatar)
	fmt.Printf("Likes: %d, Retweets: %d, Replies: %d\n", tweetWithMedia.Tweet.Likes, tweetWithMedia.Tweet.Retweets, tweetWithMedia.Tweet.Replies)

	if len(tweetWithMedia.Photos) > 0 {
		fmt.Printf("Photos:\n")
		for _, photo := range tweetWithMedia.Photos {
			fmt.Printf(" - %s\n", photo)
		}
	}

	if len(tweetWithMedia.Videos) > 0 {
		fmt.Printf("Videos:\n")
		for _, video := range tweetWithMedia.Videos {
			fmt.Printf(" - %s\n", video)
		}
	}

	if len(tweetWithMedia.GIFs) > 0 {
		fmt.Printf("GIFs:\n")
		for _, gif := range tweetWithMedia.GIFs {
			fmt.Printf(" - %s\n", gif)
		}
	}
}
