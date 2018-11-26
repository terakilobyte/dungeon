package chatbot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gempir/go-twitch-irc"
	"github.com/joho/godotenv"
	"github.com/martinlindhe/notify"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/terakilobyte/dungeon/chatbot/commands"
)

var chatClient ChatClient

// ChatClient reprsents a chatbot chat client
type ChatClient struct {
	Client     *twitch.Client
	Collection *mongo.Collection
}

// GetClient gets a chat client
func GetClient() (*ChatClient, error) {
	if chatClient.Client == nil || chatClient.Collection == nil {
		return nil, fmt.Errorf("chat client isn't instantiated")
	}
	return &chatClient, nil
}

func newChatClient(c *twitch.Client, m *mongo.Collection) {
	chatClient = ChatClient{Client: c, Collection: m}
}

// Run runs the chatbot
func Run() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("poop")
	}
	mongoClient, err := mongo.NewClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	collection := mongoClient.Database("twitch").Collection("viewers")
	err = mongoClient.Connect(context.TODO())
	twitchOauth := os.Getenv("TWITCH_OAUTH")
	twitchChannel := os.Getenv("TWITCH_ACCOUNT")
	client := twitch.NewClient(twitchChannel, twitchOauth)
	commandHandler := commands.NewCommand(client, twitchChannel)

	client.OnNewUnsetMessage(func(foobar string) {
		fmt.Println("Got an unset message")
		fmt.Println(foobar)
	})

	client.OnNewMessage(func(channel string, user twitch.User, message twitch.Message) {
		if user.Username == "swarmlogic_bot" {
			return
		}
		if message.Text[:1] == "!" {
			rawCommand := commands.RawCommand{
				Collection: collection,
				Args:       message.Text[1:],
				Channel:    channel,
				User:       user,
			}
			commandHandler.HandleCommand(rawCommand)
			return
		}
		notify.Notify("chatbot", fmt.Sprintf("%s said:", user.Username), message.Text, "")
	})

	client.OnUserJoin(func(channel string, user string) {
		notify.Notify("chatbot", fmt.Sprintf("%s joined the channel", user), fmt.Sprintf("Welcome %s", user), "")
		//		filter := bson.D{{"user", user}}
		//		timeJoined := bson.D{{"timeJoined", time.Now()}}
		//		opts := options.Update()
		//		opts.SetUpsert(true)
		//		collection.UpdateOne(context.Background(),
		//			filter,
		//			bson.D{{"$set", timeJoined}},
		//			opts,
		//		)
	})
	client.OnUserPart(func(channel string, user string) {
		notify.Notify("chatbot", fmt.Sprintf("%s left the channel", user), "Bye bye!", "")
		//		res := &bson.D{}
		//		elem := collection.FindOne(context.Background(), bson.D{{"user", user}})
		//		if err := elem.Decode(res); err != nil {
		//			fmt.Println(err, "error decoding document")
		//		}
		//		doc := res.Map()
		//		totalTime := 0 * time.Second
		//		if val, ok := doc["totalTime"]; ok {
		//			totalTime += time.Duration(val.(int64))
		//		}
		//		timeJoined := time.Unix(int64(doc["timeJoined"].(primitive.DateTime)/1000), 1000)
		//		timeElapsedThisSession := time.Since(timeJoined)
		//		update := bson.D{{"user", user}, {"totalTime", timeElapsedThisSession + totalTime}}
		//		opts := options.Update()
		//		opts.SetUpsert(true)
		//		fmt.Printf("%+v\n", update)
		//		_, err := collection.UpdateOne(context.Background(),
		//			bson.D{{"user", user}},
		//			bson.D{{"$set", update}},
		//			opts,
		//		)
		//		if err != nil {
		//			fmt.Println("error updating totalTime", err)
		//		}

	})

	client.OnConnect(func() {
		notify.Notify("chatbot", "chatbot", "chatbot online", "")
		fmt.Println("Connected to the channel, bot online")
		newChatClient(client, collection)
	})

	client.Join(twitchChannel)
	if err := client.Connect(); err != nil {
		panic(err)
	}
}
