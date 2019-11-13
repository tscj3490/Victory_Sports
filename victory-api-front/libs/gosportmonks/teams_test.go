package gosportmonks

import (
	"reflect"
	"testing"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
	"fmt"
)

//var (
//	DraftDB = db.TestDB
//	ctx = context.Background()
//	netTransport = &http.Transport{
//		Dial: (&net.Dialer{
//			Timeout: 5 * time.Second,
//		}).Dial,
//		TLSHandshakeTimeout: 5 * time.Second,}
//	netClient = &http.Client{
//		Timeout: time.Second * 10,
//		Transport: netTransport,
//	}
//	client = NewClient(netClient)
//)

func TestTeamsServiceOp_List(t *testing.T) {
	type fields struct {
		client *Client
	}
	type args struct {
		ctx      context.Context
		seasonID uint
		opt      *ListOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Team
		want1   *Response
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TeamsServiceOp{
				client: tt.fields.client,
			}
			got, got1, err := s.List(tt.args.ctx, tt.args.seasonID, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("TeamsServiceOp.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TeamsServiceOp.List() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("TeamsServiceOp.List() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTeamsServiceOp_Get(t *testing.T) {
	teamIds := map[string]uint{
		"Al Ain": 7780,
	}

	type fields struct {
		client *Client
	}
	type args struct {
		ctx    context.Context
		teamID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Team
		want1   *Response
		wantErr bool
	}{
		{
			fields: fields{
				client: client,
			},
			args: args{
				ctx: ctx,
				teamID: teamIds["Al Ain"],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TeamsServiceOp{
				client: tt.fields.client,
			}
			got, got1, err := s.Get(tt.args.ctx, tt.args.teamID)
			fmt.Printf("got %v\n", got)
			if (err != nil) != tt.wantErr {
				fmt.Printf("req: %v", got1)
				t.Errorf("TeamsServiceOp.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("TeamsServiceOp.Get() got = %v, want %v", got, tt.want)
			//}
			//if !reflect.DeepEqual(got1, tt.want1) {
			//	t.Errorf("TeamsServiceOp.Get() got1 = %v, want %v", got1, tt.want1)
			//}
		})
	}
}
