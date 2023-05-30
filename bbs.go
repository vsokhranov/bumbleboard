package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	MAX_POSTS       int = 100
	MAX_POST_LENGTH int = 400
	POSTS_PER_IP    int = 10
)

type Post struct {
	Time    time.Time `json:"time"`
	Name    string    `json:"name"`
	Content string    `json:"content"`
	IP      string    `json:"ip"`
}

// BBS is the main struct for the BBS.
// It contains a slice of posts and a map of IP addresses to their post counts
type BBS struct {
	Posts        []*Post
	Mutex        sync.Mutex
	PostsCounter map[string]int
	PostsLogger  *log.Logger
}

var (
	// Message from admins to display inside text area
	topAlert string = fmt.Sprintf(
		"Old posts are deleted every 7 days. %d posts up to %d characters per user daily.",
		POSTS_PER_IP, MAX_POST_LENGTH)

	// Custom template functions
	funcs = template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
)

// FlushPosts removes all posts from the BBS except the last MAX_POSTS
func (bbs *BBS) FlushPosts() {
	bbs.Mutex.Lock()
	defer bbs.Mutex.Unlock()
	bbs.Posts = bbs.Posts[len(bbs.Posts)-1:]
	log.Print("Flushed old posts…")
}

// RefreshLimit refreshes the limit of posts per IP
func (bbs *BBS) RefreshLimit() {
	bbs.Mutex.Lock()
	defer bbs.Mutex.Unlock()
	bbs.PostsCounter = make(map[string]int)
	log.Print("Refreshed limit of posts per IP…")
}

// PostCount returns the number of posts by an IP
func (bbs *BBS) PostCount(ip string) int {
	return bbs.PostsCounter[ip]
}

// AddPost adds a new post
func (bbs *BBS) AddPost(content string, ip string) {
	bbs.Mutex.Lock()
	defer bbs.Mutex.Unlock()
	// Flush old posts if there are more than MAX_POSTS
	if len(bbs.Posts) >= MAX_POSTS {
		bbs.FlushPosts()
	}
	if bbs.PostsCounter[ip] >= POSTS_PER_IP {
		return
	}

	post := &Post{
		Time:    time.Now(),
		Name:    nameFromIP(ip),
		Content: content,
		IP:      ip,
	}

	bbs.Posts = append(bbs.Posts, post)
	bbs.PostsLogger.Printf("%s: %s", post.Name, post.Content)
	bbs.PostsCounter[ip]++
}

// Save saves the BBS posts to a file
func (bbs *BBS) SavePosts() {
	log.Print("Saving posts to posts.json…")
	if _, err := os.Stat("posts.json"); os.IsNotExist(err) {
		log.Print("Creating posts.json file…")
		file, err := os.Create("posts.json")
		if err != nil {
			log.Printf("Error creating posts.json file! %s", err)
		}
		defer file.Close()
	}
	bbs.Mutex.Lock()
	defer bbs.Mutex.Unlock()
	jsonBytes, err := json.Marshal(bbs.Posts)
	if err != nil {
		log.Printf("Error marshalling posts.json file! %s", err)
	}
	err = os.WriteFile("posts.json", jsonBytes, 0644)
	if err != nil {
		log.Printf("Error writing posts.json file! %s", err)
	}
	log.Print("Successfully wrote posts.json file!")
}

// LoadPosts loads the BBS posts from a file
func (bbs *BBS) LoadPosts() {
	bbs.Mutex.Lock()
	defer bbs.Mutex.Unlock()
	file, err := os.Open("posts.json")
	if err != nil {
		log.Printf("Error opening posts.json file: %s", err)
		return
	}
	defer file.Close()
	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading posts.json file: %s", err)
	}
	err = json.Unmarshal(jsonBytes, &bbs.Posts)
	if err != nil {
		log.Printf("Error unmarshalling posts.json file: %s", err)
	}
	log.Print("Successfully loaded posts.json file!")
}

// getNameArrays returns two arrays of names for adjectives and animals.
// This is used for generating random names for post authors.
// Using separate function is better than using a global variable.
func getNameArrays() ([]string, []string) {
	return []string{
			"happy", "sad", "angry", "tired", "sleepy",
			"hungry", "thirsty", "quiet", "loud", "fast",
			"slow", "dark", "light", "bright", "dim",
			"brave", "timid", "joyful", "somber", "optimistic",
			"pessimistic", "friendly", "unfriendly", "shy", "outgoing",
			"adventurous", "cautious", "clever", "foolish", "serious",
			"playful", "loving", "hateful", "jealous", "envious",
			"generous", "selfish", "kind", "cruel", "polite",
			"rude", "honest", "dishonest", "brave", "fearful",
			"anxious", "calm", "peaceful", "careful", "careless",
			"clumsy", "graceful", "curious", "boring", "entertaining",
			"fascinating", "intriguing", "mysterious", "intelligent", "dumb",
			"talented", "untalented", "creative", "uncreative", "patient",
			"impatient", "relaxed", "tense", "messy", "neat",
			"organized", "disorganized", "hopeful", "hopeless", "romantic",
			"unromantic", "charming", "dull", "stylish", "unstylish",
			"sophisticated", "unsophisticated", "beautiful", "ugly", "handsome",
			"pretty", "stunning", "ordinary", "unique", "powerful",
			"weak", "lively", "boring", "trustworthy", "untrustworthy",
			"faithful", "unfaithful", "independent", "dependent", "confident"},
		[]string{
			"lion", "tiger", "bear", "elephant", "giraffe",
			"hippopotamus", "rhinoceros", "gorilla", "chimpanzee", "orangutan",
			"zebra", "cheetah", "jaguar", "leopard", "cougar",
			"lynx", "bobcat", "panther", "wolf", "coyote",
			"fox", "hyena", "badger", "raccoon", "skunk",
			"otter", "beaver", "squirrel", "chipmunk", "rabbit",
			"hare", "deer", "moose", "elk", "caribou",
			"bison", "buffalo", "yak", "camel", "llama",
			"alpaca", "kangaroo", "koala", "platypus", "shark",
			"wombat", "ostrich", "emu", "penguin", "seagull",
			"pelican", "flamingo", "parrot", "toucan", "hummingbird",
			"bald eagle", "falcon", "osprey", "hawk", "sparrow",
			"dove", "pigeon", "crow", "raven", "magpie",
			"blue jay", "cardinal", "robin", "bluebird", "woodpecker",
			"kingfisher", "bee", "butterfly", "caterpillar", "ladybug",
			"ant", "spider", "scorpion", "centipede", "millipede",
			"snake", "python", "anaconda", "cobra", "rattlesnake",
			"garter snake", "turtle", "tortoise", "alligator", "crocodile",
			"frog", "toad", "salamander", "newt", "jellyfish",
			"crab", "lobster", "shrimp", "dolphin", "whale"}
}

// nameFromIP generates a name from an IP address.
// The name is generated by concatenating the random adjective and animal name.
func nameFromIP(ip string) string {
	// Get two arrays of names for adjectives and animals
	adjectives, animals := getNameArrays()
	// Parse the IP address string
	parsedIP := net.ParseIP(ip)
	// Compute the SHA256 hash of the IP address
	hash := sha256.Sum256(parsedIP)
	// Split the hex string into two parts and use them as indices into the adjective and animal arrays
	adjectiveIndex := int(hash[0]) % len(adjectives)
	animalIndex := int(hash[1]) % len(animals)
	// Combine the adjective and animal name into a single string
	return adjectives[adjectiveIndex] + " " + animals[animalIndex]
}

// sanitizePost is a helper function to sanitize user input
func sanitizePost(text string) string {
	text = strings.TrimSpace(text)
	if len(text) > MAX_POST_LENGTH {
		text = text[:MAX_POST_LENGTH]
	}
	text = html.EscapeString(text)
	return text
}

func main() {
	// Create a new HTTP server
	srv := &http.Server{
		Addr:    ":8081",
		Handler: nil, // use the default ServeMux
	}

	// Listen for signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Open file for appending
	logfile, err := os.OpenFile("posts.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	// Our BBS instance
	bbs := &BBS{
		Posts:        make([]*Post, 0, MAX_POSTS),
		PostsCounter: make(map[string]int),
		PostsLogger:  log.New(logfile, "", log.LstdFlags),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}
		if r.Method == "POST" {
			content := sanitizePost(r.FormValue("content"))
			if content == "" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			bbs.AddPost(content, ip)
			log.Printf("%s (%s) said: %s", nameFromIP(ip), ip, content)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		tpl := template.Must(template.New("index.html").Funcs(funcs).ParseFiles("index.html"))
		data := struct {
			Posts     []*Post
			PostsLeft int
			TopAlert  string
		}{
			Posts:     bbs.Posts,
			PostsLeft: POSTS_PER_IP - bbs.PostCount(ip),
			TopAlert:  topAlert,
		}
		tpl.Execute(w, data)

	})

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Delete old posts every week
	go func() {
		for range time.Tick(168 * time.Hour) {
			bbs.FlushPosts()
		}
	}()

	// Refresh limit every 24 hours
	go func() {
		for range time.Tick(24 * time.Hour) {
			bbs.RefreshLimit()
		}
	}()

	// Start the server
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		bbs.LoadPosts()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s", err)
		}
	}()

	// Wait for a signal to shutdown the server
	<-stop

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	bbs.SavePosts()
	log.Println("Server stopped")
}
