// Copyright 2019 Citra Digital Lintas
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"errors"
	io "io"
	"log"
	"os"
	"testing"

	"github.com/citradigital/protonats"
	"github.com/stretchr/testify/assert"
)

type TestProtonatsService struct {
	Fixtures Fixtures
}

func (b *TestProtonatsService) GetTestA(ctx context.Context, req *TestARequest) (*TestAResponse, error) {
	if req.Input == "123456" {
		return nil, errors.New("test-error-1")
	}

	id := ctx.Value(string("BusID"))

	if id != nil {
		b.Fixtures.SetCounter(id.(string))
	}
	result := &TestAResponse{
		Output: "OK" + req.Input,
		Id:     req.Id,
	}
	return result, nil
}

func (b *TestProtonatsService) FeedData(stream TestService_FeedDataProtonatsServer) {
	var sum int64

	var data *FeedDataRequest
	var err error
	for {
		data, err = stream.Receive()
		if b.Fixtures != nil && b.Fixtures.GetValue() == "crash" {
			err = errors.New("crash")
		}

		if err != nil {
			break
		}

		sum = sum + data.Data
	}

	if b.Fixtures != nil && b.Fixtures.GetValue() == "crash2" {
		err = errors.New("crash2")
	}

	if err == io.EOF {
		err := stream.Done(&FeedDataResponse{Sum: sum})

		if err != nil {
			stream.Error(err)
		}
	} else if err != nil {
		stream.Error(err)
	}

}

func (b *TestProtonatsService) StreamData(req *StreamDataRequest, stream TestService_StreamDataProtonatsServer) error {
	// We have a set of data which will be multiplied by the req
	// and stream those numbers down to the client
	data := [10]int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	for i := range data {
		if b.Fixtures != nil && b.Fixtures.GetValue() == "crash" {
			return errors.New("crash")
		}
		err := stream.Send(&StreamDataResponse{Data: data[i] * req.Id})

		if err != nil {
			return err
		}
	}
	return nil
}

func createTestService() *TestProtonatsService {
	test := TestProtonatsService{
		Fixtures: CreateFixtures(),
	}

	return &test
}

var natsURL string

func TestInit(t *testing.T) {
	natsURL = os.Getenv("NATS_URL")
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
}

func TestError1(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()

	svr := NewTestServiceProtonatsServer(bus, d)
	_, err = svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	client, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)

	_, err = svc.GetTestA(ctx, &TestARequest{Input: "123456"})

	assert.NotEqual(t, nil, err)
	assert.Equal(t, "test-error-1", err.Error())
}

func TestOK1(t *testing.T) {
	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	resp, err := svc.GetTestA(ctx, &TestARequest{Input: "OK"})

	assert.Equal(t, nil, err)
	assert.Equal(t, "OKOK", resp.Output)

	cancel()
	<-done
}

/*
func TestOKLoop(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL, ID: "bus1"})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	_, err = svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	bus2, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL, ID: "bus2"})
	assert.Equal(t, nil, err)
	defer bus2.Close()
	svr2 := NewTestServiceProtonatsServer(bus2, d)
	_, err = svr2.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	max := 100000
	svc := NewTestServiceProtonatsClient(client)

	t1 := time.Now()
	for i := 0; i < max; i++ {
		resp, err := svc.GetTestA(ctx, &TestARequest{Input: "OK", Id: int64(i)})

		if err != nil {
			t.Fail()
		}
		if resp.Output != "OKOK" {
			t.Fail()
		}
		if i%10000 == 0 {
			log.Println(i)
		}
	}
	t2 := time.Now()

	dur := t2.Sub(t1).Seconds()
	log.Printf("%f reqs/sec\n", float64(max)/dur)
	assert.Equal(t, true, (d.Fixtures.GetCounter("bus1") < max))
	assert.Equal(t, true, (d.Fixtures.GetCounter("bus2") < max))

	log.Println(d.Fixtures)
}
*/

func TestClientStreamHappy(t *testing.T) {
	log.Println("ClientStreamHappy")
	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	stream, err := svc.FeedData(ctx)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	for i := 0; i < 10; i++ {
		_ = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})
	}

	resp, err := stream.Done()

	assert.Equal(t, int64(45), resp.Sum)
	cancel()
	<-done
}

func TestClientStreamSad1(t *testing.T) {
	log.Println("ClietnStreamSad1")
	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	stream, err := svc.FeedData(ctx)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	for i := 0; i < 10; i++ {
		if i == 7 {
			// simulate crash on 7th iteration
			d.Fixtures.SetValue("crash")
		}
		err = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})

		if err != nil {
			assert.NotEqual(t, nil, err)
			break
		}
	}

	resp, err := stream.Done()

	assert.NotEqual(t, nil, err)

	assert.Equal(t, true, resp == nil)
	cancel()
	<-done
}

func TestClientStreamSad2(t *testing.T) {
	log.Println("ClietnStreamSad2")
	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	stream, err := svc.FeedData(ctx)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	// simulate crash
	d.Fixtures.SetValue("crash2")

	for i := 0; i < 10; i++ {
		err = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})

		if err != nil {
			assert.NotEqual(t, nil, err)
			break
		}
	}

	resp, err := stream.Done()

	assert.NotEqual(t, nil, err)

	assert.Equal(t, true, resp == nil)
	cancel()
	<-done
}

func TestServerStreamHappy(t *testing.T) {
	log.Println("ServerStreamHappy")

	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	stream, err := svc.StreamData(ctx, &StreamDataRequest{
		Id: 2,
	})

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	count := 0
	var sum int64
	for {
		// Wait for the data to be available from the stream
		data, err := stream.Receive()
		if count == 10 {
			assert.Equal(t, io.EOF, err)
		} else {
			assert.Equal(t, nil, err)
		}
		if err != nil {
			break
		}
		sum = sum + data.Data
		count++
	}

	assert.Equal(t, int64(110), sum)
	assert.Equal(t, 10, count)

	cancel()
	<-done
}

func TestServerStreamSad1(t *testing.T) {
	log.Println("ServerStreamSad")

	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	stream, err := svc.StreamData(ctx, &StreamDataRequest{
		Id: 2,
	})

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	count := 0
	var sum int64
	for {
		// Wait for the data to be available from the stream
		data, err := stream.Receive()
		if count == 8 {
			assert.Equal(t, errors.New("crash"), err)
		} else {
			assert.Equal(t, nil, err)
		}
		if err != nil {
			break
		}
		sum = sum + data.Data
		count++
		if count == 7 {
			d.Fixtures.SetValue("crash")
		}
	}

	assert.Equal(t, int64(104), sum)
	assert.Equal(t, 8, count)

	cancel()
	<-done
}
