package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/googollee/go-socket.io"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/sportmonks_api"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
)

const (
	room     = "chat"
	listener = "livescore:update"
	msgTest  = "test"
)

var livescoresAlreadyRunning = false

var (
	netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second}
	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	client = gosportmonks.NewClient(netClient)
)

// list of socket.io clients
var clients = make(map[string]socketio.Socket)

func tester(ctx context.Context, c chan string) {
	if livescoresAlreadyRunning {
		livescoresAlreadyRunning = false
		// league options:
		liveOpt := &gosportmonks.ListOptions{
			Include: "events,inplay,localTeam,visitorTeam,substitutions,goals,cards,other,corners,lineup,bench,sidelined,stats,comments,tvstations",
		}
		livescorefixtures, _, err := client.Livescores.ListNow(ctx, liveOpt)
		if err != nil {
			log.Printf("Livescores Errors: %v", err)
		}
		if len(livescorefixtures) > 0 {
			var testFixture gosportmonks.LivescoreFixture
			testFixture = livescorefixtures[0]
			log.Printf("Livescores: %v %v", testFixture.ID, testFixture.EventsInclude)
		}
<<<<<<< HEAD
		time.Sleep(time.Duration(30 * time.Second))
		log.Println("Starting test...")
		for i := 0; i < 10; i++ {
			time.Sleep(time.Duration(10 * time.Millisecond))
			ret, _ := json.Marshal(livescorefixtures)
			c <- string(ret)
		}
=======

		smAPI := &sportmonks_api.SportmonksAPI{
			Ctx:                    ctx,
			RebuildCache:           false,
			DontTriggerCacheDBSave: false,
		}
		// convert competions into hashmap for easy lookup
		compExists := map[uint]bool{}
		for _, c := range smAPI.ListCompetitions() {
			compExists[c.ID] = true
		}
		// added a filter to include only currently included leagues
		tmpFixtures := []gosportmonks.LivescoreFixture{}
		for _, f := range livescorefixtures {
			if _, ok := compExists[f.LeagueID]; !ok {
				continue
			}
			tmpFixtures = append(tmpFixtures, f)
		}

		// time.Sleep(time.Duration(50 * time.Second))
		log.Println("Starting test...")
		// for i := 0; i < 10; i++ {
		// time.Sleep(time.Duration(10 * time.Millisecond))
		ret, _ := json.Marshal(tmpFixtures)
		c <- string(ret)
		// }
>>>>>>> 4fa22ea2c95fb7153a2c7841ddfb6378a3c883d5
		log.Println("Ended test.")
		livescoresAlreadyRunning = true
	}

}

// listen for messages on channel and broadcast to the room
func broadcast(c chan string, sioRoom, msgType string) {
	for {
		msg := <-c
		for _, clientSIO := range clients {
			clientSIO.BroadcastTo(sioRoom, msgType, msg)
		}
	}
}

// Livescores controller implementes the livescores gosportmonks endpoint to allow subscribers to get updated livescores via websockets
type Livescores struct {
	server *socketio.Server
	ch     chan string
}

// OnError the socketio error handler
func (l *Livescores) OnError(so socketio.Socket, err error) {
	log.Println("error:", err)
}

// OnConnection starts the hook in the socketio server
func (l *Livescores) OnConnection(so socketio.Socket) {
	so.On("disconnection", func() {
		log.Println("on disconnect")
	})
	id := so.Id()
	clients[id] = so
	so.Join(room)
	msg := "hello"
	l.server.BroadcastTo(room, "chat message", msg)
	ctx := context.Background()

	go broadcast(l.ch, room, listener)
	// go tester(ctx, l.ch)
	ticker := time.NewTicker(30 * time.Second)

	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				tester(ctx, l.ch)
			case <-quit:
				ticker.Stop()
				return
			}
			time.Sleep(time.Duration(1 * time.Second))
		}
	}()

	log.Println("on connection")
}
