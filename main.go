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

var api = gotwitch.New(os.Getenv("CLIENT_ID"))

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

type channelStats struct {
	ChannelName string

	EmoteCount int
}

type emoteStats struct {
	EmoteName string

	// Total usage of the emote in the last 60 seconds
	EmoteCount int

	// Participating channels
	Channels []channelStats
}

// Stats xD
type Stats struct {
	EmoteStats []emoteStats
}

func printStats() {
	for _ = range time.Tick(2 * time.Second) {
		stats := Stats{}
		recordKeeper.mutex.Lock()
		for emoteName, emoteRecord := range recordKeeper.Records {
			emoteStat := emoteStats{}
			emoteStat.EmoteName = emoteName
			emoteRecord.mutex.Lock()
			if len(emoteRecord.channelEmoteCount) == 0 {
				emoteRecord.mutex.Unlock()
				continue
			}
			sorted := sortedResults(emoteRecord.channelEmoteCount)
			emoteRecord.mutex.Unlock()
			// fmt.Printf("Current record for %s is %#v\n", emoteName, emoteRecord)
			for _, channel := range sorted {
				channelStat := channelStats{}
				channelStat.ChannelName = channel.Key
				channelStat.EmoteCount = channel.Value

				emoteStat.EmoteCount += channelStat.EmoteCount

				emoteStat.Channels = append(emoteStat.Channels, channelStat)
			}
			// Figure out which channel is in the lead
			// Sort emote
			stats.EmoteStats = append(stats.EmoteStats, emoteStat)
		}
		recordKeeper.mutex.Unlock()

		sort.Slice(stats.EmoteStats, func(i, j int) bool {
			// return stats.EmoteStats[i].Channels[0].EmoteCount < stats.EmoteStats[j].Channels[0].EmoteCount
			return stats.EmoteStats[i].EmoteCount > stats.EmoteStats[j].EmoteCount
		})

		fmt.Print("\033[2J")
		for emoteRank, emote := range stats.EmoteStats {
			fmt.Printf("Emote %s stats(%d):\n", emote.EmoteName, emote.EmoteCount)
			for channelRank, channel := range emote.Channels {
				fmt.Printf("#%d: %s with %d emotes\n", channelRank+1, channel.ChannelName, channel.EmoteCount)
				if channelRank >= 2 {
					break
				}
			}
			fmt.Println("")

			if emoteRank >= 4 {
				break
			}
		}
	}
}

func init() {
	recordKeeper = NewRecordKeeper()
}

func main() {
	go printStats()

	client := twitch.NewClient("justinfan123123", "oauth:123123123")

	client.OnNewMessage(func(channelName string, user twitch.User, message twitch.Message) {
		// TODO: Parse BTTV emotes, FFZ emotes, and Emojis here
		if len(message.Emotes) == 0 {
			// Ignore messages with no emotes
			return
		}

		recordKeeper.GetEmoteRecord(message.Emotes[0].Name).IncrementCounter(channelName)
	})

	api.GetStreams(func(streams []gotwitch.Stream) {
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
