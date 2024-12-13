package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

func simulateUsers(engine *Engine, numUsers int, numSubReddits int) {
	// Create subreddits
	for i := 0; i < numSubReddits; i++ {
		subRedditName := fmt.Sprintf("SubReddit%d", i+1)
		engine.CreateSubReddit(subRedditName)
	}

	for i := 0; i < numUsers; i++ {
		username := fmt.Sprintf("User%d", i+1)
		user := engine.RegisterUser(username)
		subCount := int(float64(numSubReddits)*math.Pow(rand.Float64(), 1.2)) + 1

		// Join random subreddits
		for j := 0; j < subCount && j < numSubReddits; j++ {
			subRedditName := fmt.Sprintf("SubReddit%d", j+1)
			engine.JoinSubReddit(user, subRedditName)
		}

		// Randomly disconnect/connect users
		if rand.Float64() > 0.2 {
			user.Connected = true
		} else {
			user.Connected = false
			engine.DisconnectedUsers++
		}

		// Create posts and comments
		for j := 0; j < rand.Intn(3)+1; j++ {
			subRedditName := fmt.Sprintf("SubReddit%d", rand.Intn(numSubReddits)+1)
			post := engine.CreatePost(user, subRedditName, fmt.Sprintf("Post content %d from %s", j+1, username))
			if post != nil {
				// Simulate random upvotes and downvotes for posts
				for k := 0; k < rand.Intn(5)+1; k++ {
					if rand.Float64() < 0.7 {
						engine.UpvotePost(post)
					} else {
						engine.DownvotePost(post)
					}
				}

				// Simulate comments on posts
				for l := 0; l < rand.Intn(2)+1; l++ {
					comment := engine.CommentPost(user, post, fmt.Sprintf("Comment %d on post %d", l+1, post.ID))

					// Simulate random upvotes and downvotes on comments
					for v := 0; v < rand.Intn(5)+1; v++ {
						if rand.Float64() < 0.7 { // 70% chance to upvote
							comment.Votes++
							comment.Author.Karma++
							engine.TotalVotes++
							engine.TotalUpvotes++
						} else { // 30% chance to downvote
							comment.Votes--
							comment.Author.Karma--
							engine.TotalVotes++
							engine.TotalDownvotes++
						}
					}

					// Simulate replies to comments
					for m := 0; m < rand.Intn(2)+1; m++ {
						engine.AddReplyToComment(user, comment, fmt.Sprintf("Reply %d to comment %d", m+1, comment.ID))
					}
				}

				// Simulate reposts
				if rand.Float64() < 0.1 {
					engine.CreateRepost(user, post, fmt.Sprintf("SubReddit%d", rand.Intn(numSubReddits)+1))
				}
			}
		}

		// Simulate direct messages
		if rand.Float64() < 0.2 && len(engine.Users) > 1 {
			targetUserID := rand.Intn(len(engine.Users)) + 1
			if targetUserID != user.ID {
				targetUser := engine.Users[targetUserID]
				engine.SendDirectMessage(user, targetUser, fmt.Sprintf("Hello from %s to %s!", user.Username, targetUser.Username))
			}
		}
	}
}

func printComments(comments []Comment, level int) {
	indent := strings.Repeat("  ", level)
	for i := range comments {
		comment := &comments[i]

		if comment.Votes == 0 {
			for v := 0; v < rand.Intn(5)+1; v++ {
				if rand.Float64() < 0.7 {
					comment.Votes++
					comment.Author.Karma++
				} else {
					comment.Votes--
					comment.Author.Karma--
				}
			}
		}

		fmt.Printf("%sComment ID %d by %s: %s (Votes: %d)\n", indent, comment.ID, comment.Author.Username, comment.Content, comment.Votes)

		if len(comment.Replies) > 0 {
			printComments(comment.Replies, level+1)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	engine := NewEngine()

	// Simulate users and subreddits
	numUsers := 100
	numSubReddits := 10
	simulateUsers(engine, numUsers, numSubReddits)

	// Calculate throughput
	duration := time.Since(engine.StartTime).Seconds()
	throughput := float64(engine.TotalActions) / duration

	// Initialize upvote and downvote counters
	totalUpvotes := 0
	totalDownvotes := 0

	for _, subreddit := range engine.SubReddits {
		for _, post := range subreddit.Posts {
			if post.Votes > 0 {
				totalUpvotes += post.Votes
			} else {
				totalDownvotes += -post.Votes
			}
		}
	}

	fmt.Println("Simulation Complete. Metrics:")
	fmt.Printf("Users: %d\n", len(engine.Users))
	fmt.Printf("SubReddits: %d\n", len(engine.SubReddits))
	fmt.Printf("Total Posts: %d\n", engine.TotalPosts)
	fmt.Printf("Total Votes: %d (Upvotes: %d, Downvotes: %d)\n", engine.TotalVotes, totalUpvotes, totalDownvotes)
	fmt.Printf("Total Comments: %d\n", engine.TotalComments)
	fmt.Printf("Total Messages: %d\n", engine.TotalMessages)
	fmt.Printf("Total Actions: %d\n", engine.TotalActions)
	fmt.Printf("Throughput (actions/sec): %.2f\n", throughput)
	fmt.Printf("Disconnected Users: %d\n", engine.DisconnectedUsers)

	// Display Action Breakdown
	fmt.Println("\nAction Breakdown:")
	for action, count := range engine.ActionBreakdown {
		fmt.Printf("%s: %d\n", action, count)
	}

	// Display Subreddit Metrics
	fmt.Println("\nSubReddit Metrics (Zipf Distribution Impact):")
	type SubRedditStats struct {
		Name      string
		Members   int
		PostCount int
	}
	var subredditStats []SubRedditStats
	for name, subreddit := range engine.SubReddits {
		stats := SubRedditStats{
			Name:      name,
			Members:   len(subreddit.Users),
			PostCount: len(subreddit.Posts),
		}
		subredditStats = append(subredditStats, stats)
	}

	sort.Slice(subredditStats, func(i, j int) bool {
		return subredditStats[i].Members > subredditStats[j].Members
	})

	for i, stats := range subredditStats {
		fmt.Printf("%d. %s - Members: %d, Posts: %d\n", i+1, stats.Name, stats.Members, stats.PostCount)
	}

	// Display Top Users by Karma
	fmt.Println("\nTop Users by Karma:")
	type UserStats struct {
		Username string
		Karma    int
	}
	var userStats []UserStats
	for _, user := range engine.Users {
		userStats = append(userStats, UserStats{Username: user.Username, Karma: user.Karma})
	}

	sort.Slice(userStats, func(i, j int) bool {
		return userStats[i].Karma > userStats[j].Karma
	})

	for i, stats := range userStats {
		fmt.Printf("%d. %s - Karma: %d\n", i+1, stats.Username, stats.Karma)
		if i >= 9 { // Display top 10 users only
			break
		}
	}

	// Display Random User Feed
	fmt.Println("\nFeed for a Random User:")
	randomUser := engine.Users[rand.Intn(len(engine.Users))+1]
	feed := engine.GetUserFeed(randomUser)
	for _, post := range feed {
		fmt.Printf("Post ID %d by %s: %s (Votes: %d)\n", post.ID, post.Author.Username, post.Content, post.Votes)
		if len(post.Comments) > 0 {
			fmt.Println("  Comments:")
			printComments(post.Comments, 1)
		}
	}

	// Display Direct Messages
	fmt.Println("\nDirect Messages:")
	for _, message := range engine.Messages {
		fmt.Printf("From %s to %s: %s\n", message.From.Username, message.To.Username, message.Content)
	}
}
