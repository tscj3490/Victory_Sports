package gosportmonks

import (
	"testing"

	"fmt"
	"net"
	"net/http"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
)

type LeagueIDs struct {
	WorldCup        int
	ChampionsLeague int
}

var leagueIds = LeagueIDs{
	WorldCup:        732,
	ChampionsLeague: 2,
}

var (
	DraftDB      = db.TestDB
	ctx          = context.Background()
	netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second}
	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	client = NewClient(netClient)
)

func TestLeaguesServiceOp_Get(t *testing.T) {
	type fields struct {
		client *Client
	}
	type args struct {
		ctx  context.Context
		name string
	}
	// bla
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *League
		want1   *Response
		wantErr bool
	}{
	// TODO: Add test cases.
	//	{
	//		fields: fields{
	//			client: client,
	//		},
	//		args: args{
	//			ctx: ctx,
	//			name: leagueIds.WorldCup,
	//		},
	//	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LeaguesServiceOp{
				client: tt.fields.client,
			}
			got, got1, err := s.Get(tt.args.ctx, tt.args.name)

			fmt.Printf("got %v\n", got)

			if (err != nil) != tt.wantErr {
				fmt.Printf("req: %v", got1)
				t.Errorf("LeaguesServiceOp.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("LeaguesServiceOp.Get() got = %v, want %v", got, tt.want)
			//}
			//if !reflect.DeepEqual(got1, tt.want1) {
			//	t.Errorf("LeaguesServiceOp.Get() got1 = %v, want %v", got1, tt.want1)
			//}
		})
	}
}
