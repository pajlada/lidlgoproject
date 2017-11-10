package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/dankeroni/gotwitch"
	twitch "github.com/gempir/go-twitch-irc"
)

var api = gotwitch.New(os.GetEnv("CLIENT_ID"))

// RecordKeeper xD
type RecordKeeper struct {
	Records map[string]*EmoteRecord
	mutex   *sync.Mutex
}

// NewRecordKeeper xD
func NewRecordKeeper() RecordKeeper {
	e := RecordKeeper{}

	e.Records = make(map[string]*EmoteRecord)
	e.mutex = &sync.Mutex{}

	return e
}

// GetEmoteRecord xD
func (r *RecordKeeper) GetEmoteRecord(emoteName string) *EmoteRecord {
	r.mutex.Lock()
	emoteRecord, ok := r.Records[emoteName]
	if ok {
		r.mutex.Unlock()
		return emoteRecord
	}
	r.Records[emoteName] = NewEmoteRecord()
	emoteRecord = r.Records[emoteName]
	r.mutex.Unlock()
	return emoteRecord
}

// EmoteRecord xD
type EmoteRecord struct {
	channelEmoteCount map[string]int
	mutex             sync.Mutex
}

type kv struct {
	Key   string
	Value int
}

func sortedResults(input map[string]int) []kv {
	var ss []kv

	for k, v := range input {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss
}

// NewEmoteRecord xD
func NewEmoteRecord() *EmoteRecord {
	emoteRecord := &EmoteRecord{}
	emoteRecord.channelEmoteCount = make(map[string]int)

	return emoteRecord
}

var recordKeeper RecordKeeper

const duration = 60 * time.Second

// IncrementCounter xd
func (r *EmoteRecord) IncrementCounter(channelName string) {
	r.mutex.Lock()
	r.channelEmoteCount[channelName]++
	r.mutex.Unlock()

	time.AfterFunc(duration, func() {
		r.mutex.Lock()
		r.channelEmoteCount[channelName]--
		if r.channelEmoteCount[channelName] == 0 {
			delete(r.channelEmoteCount, channelName)
		}
		r.mutex.Unlock()
	})
}

func printStats() {
	for _ = range time.Tick(2 * time.Second) {
		recordKeeper.mutex.Lock()
		for emoteName, emoteRecord := range recordKeeper.Records {
			emoteRecord.mutex.Lock()
			if len(emoteRecord.channelEmoteCount) == 0 {
				emoteRecord.mutex.Unlock()
				continue
			}
			sorted := sortedResults(emoteRecord.channelEmoteCount)
			emoteRecord.mutex.Unlock()
			// fmt.Printf("Current record for %s is %#v\n", emoteName, emoteRecord)
			for i, channel := range sorted {
				fmt.Printf("#%d: %s with %d %s\n", i+1, channel.Key, channel.Value, emoteName)
			}
			fmt.Println("+++++++++++++")
			// Figure out which channel is in the lead
			// Sort emote
		}
		recordKeeper.mutex.Unlock()
		fmt.Println("==========================")
	}
}

func init() {
	recordKeeper = NewRecordKeeper()
}

func main() {
	go printStats()

	client := twitch.NewClient("justinfan123123", "oauth:123123123")

	client.OnNewMessage(func(channelName string, user twitch.User, message twitch.Message) {
		// fmt.Printf("%s: %s\n", user.DisplayName, message.Text)
		if len(message.Emotes) == 0 {
			// Ignore messages with no emotes
			return
		}

		// Allocate channel if we haven't already
		firstEmote := message.Emotes[0]
		emoteRecord := recordKeeper.GetEmoteRecord(firstEmote.Name)
		emoteRecord.IncrementCounter(channelName)
	})

	api.GetStreams(func(streams []gotwitch.Stream) {
		// fmt.Printf("XD: %#v\n", ret)
		userIDs := []string{}
		for _, stream := range streams {
			userIDs = append(userIDs, stream.UserID)
		}

		api.GetUsers(userIDs, func(users []gotwitch.User) {
			for _, user := range users {
				fmt.Printf("Join channel with id %s and username %s\n", user.ID, user.Login)
				client.Join(user.Login)
			}
		}, func(statusCode int, statusMessage, errorMessage string) {

		}, func(err error) {

		})
	}, func(statusCode int, statusMessage, errorMessage string) {

	}, func(err error) {

	})

	client.Join("pajlada")
	client.Join("forsenlol")
	client.Join("nymn")
	client.Join("nanilul")

	err := client.Connect()
	if err != nil {
		fmt.Println(err)
	}
}
