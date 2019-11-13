package gosportmonks

import (
	"reflect"
	"testing"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
	"fmt"
)

func TestStagesServiceOp_List(t *testing.T) {
	seasonIds := map[string]uint{
		"UAELeague": 11697,
	}

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
		want    []Stage
		want1   *Response
		wantErr bool
	}{
	// TODO: Add test cases.
		{
			fields: fields{
				client: client,
			},
			args: args{
				ctx: ctx,
				seasonID: seasonIds["UAELeague"],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StagesServiceOp{
				client: tt.fields.client,
			}
			got, got1, err := s.List(tt.args.ctx, tt.args.seasonID, tt.args.opt)
			fmt.Printf("got %v\n", got)
			if (err != nil) != tt.wantErr {
				fmt.Printf("req: %v", got1)
				t.Errorf("StagesServiceOp.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("StagesServiceOp.List() got = %v, want %v", got, tt.want)
			//}
			//if !reflect.DeepEqual(got1, tt.want1) {
			//	t.Errorf("StagesServiceOp.List() got1 = %v, want %v", got1, tt.want1)
			//}
		})
	}
}

func TestStagesServiceOp_Get(t *testing.T) {
	type fields struct {
		client *Client
	}
	type args struct {
		ctx     context.Context
		stageID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Stage
		want1   *Response
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StagesServiceOp{
				client: tt.fields.client,
			}
			got, got1, err := s.Get(tt.args.ctx, tt.args.stageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("StagesServiceOp.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StagesServiceOp.Get() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("StagesServiceOp.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
