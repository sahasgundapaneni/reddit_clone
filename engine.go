package engine

import (
	"sync"
	"time"
)

// Data Structures

type User struct {
	ID        int
	Username  string
	Karma     int
	Actions   int
	Connected bool
}

type SubReddit struct {
	Name  string
	Posts []*Post
	Users map[int]*User
}

type Post struct {
	ID       int
	Author   *User
	Content  string
	Comments []Comment
	Votes    int
}

type Comment struct {
	ID      int
	Author  *User
	Content string
	Replies []Comment
	Votes   int
}

type Message struct {
	From    *User
	To      *User
	Content string
}

type Engine struct {
	Users             map[int]*User
	SubReddits        map[string]*SubReddit
	Messages          []Message
	PostID            int
	CommentID         int
	TotalPosts        int
	TotalVotes        int
	TotalUpvotes      int
	TotalDownvotes    int
	TotalMessages     int
	TotalActions      int
	TotalComments     int
	DisconnectedUsers int
	StartTime         time.Time
	Mutex             sync.Mutex
	ActionBreakdown   map[string]int
}

// Initialization and Utility Functions

func NewEngine() *Engine {
	return &Engine{
		Users:      make(map[int]*User),
		SubReddits: make(map[string]*SubReddit),
		Messages:   []Message{},
		PostID:     1,
		CommentID:  1,
		StartTime:  time.Now(),
		ActionBreakdown: map[string]int{
			"Posts":    0,
			"Comments": 0,
			"Votes":    0,
			"Messages": 0,
		},
	}
}

func (e *Engine) RegisterUser(username string) *User {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	id := len(e.Users) + 1
	user := &User{ID: id, Username: username, Karma: 0, Actions: 0, Connected: true}
	e.Users[id] = user
	return user
}

func (e *Engine) CreateSubReddit(name string) *SubReddit {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.SubReddits[name]; exists {
		return nil
	}
	subReddit := &SubReddit{Name: name, Posts: []*Post{}, Users: make(map[int]*User)}
	e.SubReddits[name] = subReddit
	return subReddit
}

func (e *Engine) JoinSubReddit(user *User, subRedditName string) bool {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	subReddit, exists := e.SubReddits[subRedditName]
	if !exists {
		return false
	}
	subReddit.Users[user.ID] = user
	user.Actions++
	e.TotalActions++
	return true
}

func (e *Engine) LeaveSubReddit(user *User, subRedditName string) bool {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	subReddit, exists := e.SubReddits[subRedditName]
	if !exists {
		return false
	}
	delete(subReddit.Users, user.ID)
	user.Actions++
	e.TotalActions++
	return true
}

func (e *Engine) CreatePost(user *User, subRedditName, content string) *Post {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	subReddit, exists := e.SubReddits[subRedditName]
	if !exists {
		return nil
	}

	post := &Post{
		ID:      e.PostID,
		Author:  user,
		Content: content,
		Votes:   0,
	}

	e.PostID++
	e.TotalPosts++
	e.ActionBreakdown["Posts"]++
	user.Actions++
	e.TotalActions++

	subReddit.Posts = append(subReddit.Posts, post)
	return post
}

func (e *Engine) CreateRepost(user *User, originalPost *Post, subRedditName string) *Post {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	subReddit, exists := e.SubReddits[subRedditName]
	if !exists {
		return nil
	}

	repost := &Post{
		ID:       e.PostID,
		Author:   user,
		Content:  originalPost.Content,
		Votes:    0,
		Comments: []Comment{},
	}

	e.PostID++
	e.TotalPosts++
	e.ActionBreakdown["Posts"]++
	user.Actions++
	e.TotalActions++

	subReddit.Posts = append(subReddit.Posts, repost)
	return repost
}

func (e *Engine) CommentPost(user *User, post *Post, content string) *Comment {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	comment := Comment{ID: e.CommentID, Author: user, Content: content, Replies: []Comment{}, Votes: 0}
	e.CommentID++
	post.Comments = append(post.Comments, comment)
	e.TotalComments++
	e.ActionBreakdown["Comments"]++
	user.Actions++
	e.TotalActions++
	return &comment
}

func (e *Engine) AddReplyToComment(user *User, parentComment *Comment, content string) *Comment {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	reply := Comment{ID: e.CommentID, Author: user, Content: content, Replies: []Comment{}, Votes: 0}
	e.CommentID++
	parentComment.Replies = append(parentComment.Replies, reply)
	e.TotalComments++
	e.ActionBreakdown["Comments"]++
	user.Actions++
	e.TotalActions++
	return &reply
}

func (e *Engine) UpvotePost(post *Post) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	post.Votes++
	post.Author.Karma++
	e.TotalVotes++
	e.TotalUpvotes++
	e.ActionBreakdown["Votes"]++
	e.TotalActions++
}

func (e *Engine) DownvotePost(post *Post) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	post.Votes--
	post.Author.Karma--
	e.TotalVotes++
	e.TotalDownvotes++
	e.ActionBreakdown["Votes"]++
	e.TotalActions++
}

func (e *Engine) SendDirectMessage(from, to *User, content string) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	message := Message{From: from, To: to, Content: content}
	e.Messages = append(e.Messages, message)
	e.TotalMessages++
	e.ActionBreakdown["Messages"]++
	from.Actions++
	e.TotalActions++
}

func (e *Engine) RetrieveMessages(user *User) []Message {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	var userMessages []Message
	for _, message := range e.Messages {
		if message.To == user {
			userMessages = append(userMessages, message)
		}
	}
	return userMessages
}

func (e *Engine) ReplyToMessage(user *User, original Message, content string) {
	e.SendDirectMessage(user, original.From, content)
}

func (e *Engine) GetUserFeed(user *User) []*Post {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	var feed []*Post
	for _, subreddit := range e.SubReddits {
		if _, subscribed := subreddit.Users[user.ID]; subscribed {
			for _, post := range subreddit.Posts {
				feed = append(feed, post)
			}
		}
	}
	return feed
}
