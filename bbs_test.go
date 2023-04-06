package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Run tests
	result := m.Run()

	// Delete log file if it exists
	if _, err := os.Stat("test.log"); !os.IsNotExist(err) {
		err = os.Remove("test.log")
		if err != nil {
			log.Fatalf("Failed to remove test.log: %s", err)
		}
	}
	os.Exit(result)
}

func TestNameFromIP(t *testing.T) {
	// Test with various inputs
	testTable := map[string]string{
		"127.0.0.1":    "trustworthy leopard",
		"::1":          "outgoing dove",
		"140.82.121.3": "graceful rhinoceros",
		"invalid-ip":   "clever spider",
	}
	for k, v := range testTable {
		if got := nameFromIP(k); got != v {
			t.Errorf("nameFromIP(%q) = %q, want %q", k, got, v)
		}
	}
}

func TestSanitizePost(t *testing.T) {
	// Test with various inputs
	testTable := map[string]string{
		"Hello world!":                    "Hello world!",
		"<script>alert('hello')</script>": "&lt;script&gt;alert(&#39;hello&#39;)&lt;/script&gt;",
		"                               ": "",
	}
	for k, v := range testTable {
		if got := sanitizePost(k); got != v {
			t.Errorf("sanitizePost(%q) = %q, want %q", k, got, v)
		}
	}

	// Test with input exceeding MAX_POST_LENGTH
	input := strings.Repeat("a", MAX_POST_LENGTH+1)
	want := strings.Repeat("a", MAX_POST_LENGTH)
	got := sanitizePost(input)
	if got != want {
		t.Errorf("sanitizePost(%s) = %s; expected %s", input, got, want)
	}
}

func TestBBS_AddPost(t *testing.T) {
	logfile, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	bbs := &BBS{
		Posts:        make([]*Post, 0, MAX_POSTS),
		PostsCounter: make(map[string]int),
		PostsLogger:  log.New(logfile, "", log.LstdFlags),
	}

	// Add posts from same IP
	for i := 0; i < POSTS_PER_IP; i++ {
		bbs.AddPost("Hello", "127.0.0.1")
	}

	// Try adding more posts from same IP
	bbs.AddPost("Hello", "127.0.0.1")

	if len(bbs.Posts) != POSTS_PER_IP {
		t.Errorf("AddPost: Expected %d posts but got %d", POSTS_PER_IP, len(bbs.Posts))
	}

	// Add posts from different IPs
	for i := 0; i < POSTS_PER_IP; i++ {
		bbs.AddPost("Hello", "127.0.0.2")
	}

	// Try adding more posts from a different IP
	bbs.AddPost("Hello", "127.0.0.2")

	if len(bbs.Posts) != POSTS_PER_IP*2 {
		t.Errorf("AddPost: Expected %d posts but got %d", POSTS_PER_IP*2, len(bbs.Posts))
	}
}

func TestBBS_FlushPosts(t *testing.T) {
	logfile, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	bbs := &BBS{
		Posts:        make([]*Post, 0, MAX_POSTS),
		PostsCounter: make(map[string]int),
		PostsLogger:  log.New(logfile, "", log.LstdFlags),
	}

	// Add some posts
	for i := 0; i < 10; i++ {
		for j := 1; j <= POSTS_PER_IP; j++ {
			ip := fmt.Sprintf("127.0.0.%d", j)
			bbs.AddPost("Hello", ip)
		}
	}

	bbs.FlushPosts()

	if len(bbs.Posts) != 1 {
		t.Errorf("FlushPosts: Expected %d posts but got %d", 1, len(bbs.Posts))
	}
}

func TestBBS_RefreshLimit(t *testing.T) {
	logfile, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	bbs := &BBS{
		Posts:        make([]*Post, 0, MAX_POSTS),
		PostsCounter: make(map[string]int),
		PostsLogger:  log.New(logfile, "", log.LstdFlags),
	}

	// Add some posts
	for i := 0; i < POSTS_PER_IP; i++ {
		bbs.AddPost("Hello", "127.0.0.1")
	}

	bbs.RefreshLimit()

	if bbs.PostCount("127.0.0.1") != 0 {
		t.Errorf("RefreshLimit: Expected post count to be 0 but got %d", bbs.PostCount("127.0.0.1"))
	}
}
