package main

const (
	restTemplate = `// Code generated by github.com/citradigital/toldata. DO NOT EDIT.
// package: {{ .Namespace }}
// source: {{ .File }}
package {{ .PackageName }}
{{ $Namespace := .Namespace }}
import (
  "encoding/json"
	"github.com/citradigital/toldata"
	context "golang.org/x/net/context"
	"net/http"
	"time"
)

func throwError(w http.ResponseWriter, message string, code int) {
	errorMessage := toldata.ErrorMessage{
	  ErrorMessage: message,
		Timestamp: time.Now().Unix(),
	}
  msg, err := json.Marshal(errorMessage)
	if err != nil {
		 http.Error(w, "{\"error-message\": \"internal-server-error\"}", http.StatusInternalServerError)
	} else {
	   http.Error(w, string(msg), code)
	}
} 

{{ range .Services }}{{ $ServiceName := .Name }}
{{ $Options := .Options }}


type {{ $ServiceName }}REST struct {
	Context context.Context
	Bus     *toldata.Bus
	Service *{{ $ServiceName }}ToldataClient
}

func New{{ $ServiceName }}REST(ctx context.Context, config toldata.ServiceConfiguration) (*{{ $ServiceName }}REST, error) {
	client, err := toldata.NewBus(ctx, config)
	if err != nil {
		return nil, err
	}

	service := {{ $ServiceName }}REST{
		Context: ctx,
		Bus:     client,
		Service: New{{ $ServiceName }}ToldataClient(client),
	}

	return &service, nil
}

func (svc *{{ $ServiceName }}REST) Install{{ $ServiceName }}Mux(mux *http.ServeMux) {


{{ range .Method }}	
{{ $InputType := .InputType }}
{{ $OutputType := .OutputType }}
{{ if or .ClientStreaming .ServerStreaming }}
{{ else  }}



  mux.HandleFunc("{{ getServiceOption $Options 99999 }}/{{ $Namespace }}/{{ $ServiceName }}/{{ .Name  }}", 
	func (w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			throwError(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var req {{ stripLastDot $InputType $Namespace }}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			throwError(w, err.Error(), http.StatusBadRequest)
			return
		}
		ip := strings.Split(r.RemoteAddr, ":")[0]
		ipaddr := &net.IPAddr{IP: net.ParseIP(ip)}
		peerInfo := &peer.Peer{Addr: ipaddr}
		ctxWithPeer := peer.NewContext(svc.Context, peerInfo)
		ret, err := svc.Service.{{ .Name }}(ctxWithPeer, &req)
		if err != nil {
			throwError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg, err := json.Marshal(ret)
		if err != nil {
			throwError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.Write(msg)
		}
	})
{{ end }}
{{ end }}
}
{{ end }}


	`

	grpcTemplate = `// Code generated by github.com/citradigital/toldata. DO NOT EDIT.
// package: {{ .Namespace }}
// source: {{ .File }}
package {{ .PackageName }}
{{ $Namespace := .Namespace }}

import (
	"io"
	"github.com/citradigital/toldata"
	context "golang.org/x/net/context"
)

// Workaround for template problem
func _eof_grpc() error {
	return io.EOF
}

{{ range .Services }}{{ $ServiceName := .Name }}
type {{ $ServiceName }}GRPC struct {
	Context context.Context
	Bus     *toldata.Bus
	Service *{{ $ServiceName }}ToldataClient
}

func New{{ $ServiceName }}GRPC(ctx context.Context, config toldata.ServiceConfiguration) (*{{ $ServiceName }}GRPC, error) {
	client, err := toldata.NewBus(ctx, config)
	if err != nil {
		return nil, err
	}

	service := {{ $ServiceName }}GRPC{
		Context: ctx,
		Bus:     client,
		Service: New{{ $ServiceName }}ToldataClient(client),
	}

	return &service, nil
}

func (svc *{{ $ServiceName }}GRPC) Close() {
	svc.Bus.Close()
}

{{ range .Method }}	

{{ $InputType := .InputType }}
{{ $OutputType := .OutputType }}
{{ if or .ClientStreaming .ServerStreaming }}
{{ if .ClientStreaming }}
func (svc *{{ $ServiceName }}GRPC) {{ .Name }}(stream {{ $ServiceName }}_{{ .Name }}Server) error {
	svrStream, err := svc.Service.{{ .Name }}(stream.Context())
	if err != nil {
		return err
	}

	for {
		isEOF := false
		data, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				isEOF = true
			} else {
				return err
			}
		}

		if data != nil {
			err = svrStream.Send(data)
			if err != nil {
				return err
			}
		}
		if isEOF {
			break
		}
	}

	resp, err := svrStream.Done()
	if err != nil {
		return err
	}
	err = stream.SendAndClose(resp)

	if err != nil {
		return err
	}

	return nil
}
{{ end }}

{{ if .ServerStreaming }}

func (svc *{{ $ServiceName }}GRPC) {{ .Name }}(req *{{ stripLastDot $InputType $Namespace }}, stream {{ $ServiceName }}_{{ .Name }}Server) error {
	svrStream, err := svc.Service.{{ .Name }}(stream.Context(), req)
	if err != nil {
		return err
	}

	for {
		data, err := svrStream.Receive()

		if err != nil {
			return err
		}
		err = stream.Send(data)
		if err != nil {
			return err
		}
	}

}
{{ end }}
{{ else }}
func (svc *{{ $ServiceName }}GRPC) {{ .Name }}(ctx context.Context, req *{{ stripLastDot $InputType $Namespace }}) (*{{ stripLastDot $OutputType $Namespace }}, error) {
	return svc.Service.{{ .Name }}(ctx, req)
}
{{ end }}
{{ end }}
{{ end }}

`

	rpcTemplate = `// Code generated by github.com/citradigital/toldata. DO NOT EDIT.
// package: {{ .Namespace }}
// source: {{ .File }}

package {{ .PackageName }}
import (
	"context"
	"errors"
   io "io"
	"github.com/gogo/protobuf/proto"
	"github.com/citradigital/toldata"
	nats "github.com/nats-io/nats.go"
)

// Workaround for template problem
func _eof() error {
	return io.EOF
}

{{ $Namespace := .Namespace }}
{{ range .Services }}{{ $ServiceName := .Name }}

type {{ .Name }}ToldataInterface interface {
	ToldataHealthCheck(ctx context.Context, req *toldata.Empty) (*toldata.ToldataHealthCheckInfo, error)

	{{ range .Method }}

{{ $InputType := .InputType }}
{{ $OutputType := .OutputType }}
	{{ if or .ClientStreaming .ServerStreaming }}
		{{ if .ServerStreaming }}
		{{ .Name }}(req *{{ stripLastDot $InputType $Namespace }}, stream {{ $ServiceName }}_{{ .Name }}ToldataServer) error
		{{ else }}
			{{ .Name }}(stream {{ $ServiceName }}_{{ .Name }}ToldataServer)
		{{ end }}
	{{ else }}
		{{ .Name }}(ctx context.Context, req *{{ stripLastDot $InputType $Namespace }}) (*{{ stripLastDot $OutputType $Namespace }}, error){{ end }}
	{{ end }}
}{{ end }}
{{ range .Services }}{{ $ServiceName := .Name }}
type {{ $ServiceName }}ToldataClient struct {
	Bus *toldata.Bus
}

type {{ $ServiceName }}ToldataServer struct {
	Bus *toldata.Bus
	Service {{ $ServiceName }}ToldataInterface
}

func New{{ $ServiceName }}ToldataClient(bus *toldata.Bus) * {{$ServiceName}}ToldataClient {
	s := &{{ $ServiceName }}ToldataClient{ Bus: bus }
	return s
}

func New{{ $ServiceName }}ToldataServer(bus *toldata.Bus, service {{ $ServiceName }}ToldataInterface) * {{$ServiceName}}ToldataServer {
	s := &{{ $ServiceName }}ToldataServer{ Bus: bus, Service: service }
	return s
}

func (service *{{ $ServiceName }}ToldataClient) ToldataHealthCheck(ctx context.Context, req *toldata.Empty) (*toldata.ToldataHealthCheckInfo, error) {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/ToldataHealthCheck"
	
	reqRaw, err := proto.Marshal(req)

	result, err := service.Bus.Connection.RequestWithContext(ctx, functionName, reqRaw)
	if err != nil {
		return nil, errors.New(functionName + ":" + err.Error())
	}

	if result.Data[0] == 0 {
		// 0 means no error
		p := &toldata.ToldataHealthCheckInfo{}
		err = proto.Unmarshal(result.Data[1:], p)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		var pErr toldata.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return nil, errors.New(pErr.ErrorMessage)
		} else {
			return nil, err
		}
	}
}



{{ range .Method }}	

{{ $InputType := .InputType }}
{{ $OutputType := .OutputType }}
{{ if or .ClientStreaming .ServerStreaming }}
type {{ $ServiceName }}_{{ .Name }}ToldataServer interface {
	{{ if .ClientStreaming }}
	Receive() (*{{ stripLastDot $InputType $Namespace }}, error)
	OnData(*{{ stripLastDot $InputType $Namespace }}) error
	Done(resp *{{ stripLastDot $OutputType $Namespace }}) error
	{{ end }}

	GetResponse() (*{{ stripLastDot $OutputType $Namespace }}, error)

	{{ if .ServerStreaming }}
	Send(*{{ stripLastDot $OutputType $Namespace }}) error
	{{ end }}
	
	TriggerEOF()
	Error(err error)
	OnExit(func())
	Exit()
}

type {{ $ServiceName }}_{{ .Name }}ToldataServerImpl struct {
	{{ if .ClientStreaming }}
	{{ end }}

	{{ if .ServerStreaming }}
	{{ end }}

	request   chan *{{ stripLastDot $InputType $Namespace }}
	isRequestClosed bool

	response chan *{{ stripLastDot $OutputType $Namespace }}
	
	cancel chan struct{}
	eof    chan struct{}
	err    chan error
	done   chan struct{}

	isEOF        bool
	isCanceled   bool

	streamErr 	error

	Context context.Context
	
}

func Create{{ $ServiceName }}_{{ .Name }}ToldataServerImpl(ctx context.Context) *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl {
	t := &{{ $ServiceName }}_{{ .Name }}ToldataServerImpl{}
	{{ if .ClientStreaming }}
	{{ end }}
	{{ if .ServerStreaming }}
	{{ end }}
	
	t.Context = ctx
	t.request = make(chan *{{ stripLastDot $InputType $Namespace }})
	t.response = make(chan *{{ stripLastDot $OutputType $Namespace }})
	t.cancel = make(chan struct{})
	t.eof = make(chan struct{})
	t.done = make(chan struct{})
	t.err = make(chan error)
	return t
}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Exit() {
	close(impl.done)
}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) OnExit(fn func()) {
	go func() {
		select {
		case <-impl.done:
			fn()
		}
	}()
}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) TriggerEOF() {
	if impl.streamErr != nil {
		return
	}
	if impl.isEOF == false {
		close(impl.eof)
		impl.isEOF = true
	}
}

{{ if .ClientStreaming }}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Receive() (*{{ stripLastDot $InputType $Namespace }}, error) {

	if impl.streamErr != nil {
		return nil, impl.streamErr
	}
	if impl.isEOF {
		return nil, io.EOF
	}

	select {
	case data := <-impl.request:
		return data, impl.streamErr
	case <-impl.cancel:
		return nil, impl.streamErr
	case <-impl.eof:
		return nil, io.EOF
	case err := <-impl.err:

		return nil, err

	}
}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) OnData(req *{{ stripLastDot $InputType $Namespace }}) error {
	if impl.streamErr != nil {
		return impl.streamErr
	}

	select {
	case err := <-impl.err:
		return err
	case impl.request <- req:
		return nil
	}
}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Done(resp *{{ stripLastDot $OutputType $Namespace }}) error {
	if impl.streamErr != nil {
		return impl.streamErr
	}

	select {
	case impl.response <- resp:
		return nil
	case err := <-impl.err:
		return err

	}
}


{{ end }}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) GetResponse() (*{{ stripLastDot $OutputType $Namespace }}, error) {
	if impl.streamErr != nil {
		return nil, impl.streamErr
	}

	select {
	case err := <-impl.err:
		{{ if .ServerStreaming }}
		impl.Exit()
		{{ end }}
		
		return nil, err

	case <-impl.cancel:
		{{ if .ServerStreaming }}
		impl.Exit()
		{{ end }}
		return nil, errors.New("canceled")

	case response := <-impl.response:
		return response, nil

		{{ if .ServerStreaming }}
	case <-impl.eof:
		impl.Exit()
		return nil, io.EOF
		{{ end }}
	}
}


{{ if .ServerStreaming }}
func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Send(req *{{ stripLastDot $OutputType $Namespace }}) error {

	if impl.isEOF {
		return io.EOF
	}

	if impl.streamErr != nil {
		return impl.streamErr
	}
	select {
	case impl.response <- req:
		return impl.streamErr

	case <-impl.cancel:
		return impl.streamErr
	case <-impl.eof:
		return io.EOF
	case err := <-impl.err:
		return err

	}
}

{{ end }}



func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Cancel() {
	if impl.isCanceled == false {
		close(impl.cancel)
		impl.isCanceled = true
	}
}


func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Error(err error) {
	impl.err <- err
	impl.streamErr = err
}

type {{ $ServiceName }}ToldataClient_{{ .Name }} struct {
	Context context.Context
	Service *{{ $ServiceName }}ToldataClient
	ID      string
}

{{ if .ClientStreaming }}

func (client *{{ $ServiceName }}ToldataClient_{{ .Name }}) Send(req *{{ stripLastDot $InputType $Namespace }}) error {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}_Send_" + client.ID
	if req == nil {
		return errors.New("empty-request")
	}
	reqRaw, err := proto.Marshal(req)
	result, err := client.Service.Bus.Connection.RequestWithContext(client.Context, functionName, reqRaw)
	if err != nil {
		return errors.New(functionName + ":" + err.Error())
	}

	if result.Data[0] == 0 {
		// 0 means no error
		return nil
	} else {
		var pErr toldata.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return errors.New(pErr.ErrorMessage)
		} else {
			return err
		}
	}
}

{{ end }}
{{ if .ServerStreaming }}

func (client *{{ $ServiceName }}ToldataClient_{{ .Name }}) Receive() (*{{ stripLastDot $OutputType $Namespace }}, error) {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}_Receive_" + client.ID
	
	result, err := client.Service.Bus.Connection.RequestWithContext(client.Context, functionName, nil)
	if err != nil {
		return nil, errors.New(functionName + ":" + err.Error())
	}

	if result.Data[0] == 0 {
		// 0 means no error
		p := &{{ stripLastDot $OutputType $Namespace }}{}
		err = proto.Unmarshal(result.Data[1:], p)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		var pErr toldata.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return nil, errors.New(pErr.ErrorMessage)
		} else {
			return nil, err
		}
	}
}
{{ end }}


func (client *{{ $ServiceName }}ToldataClient_{{ .Name }}) Done() (*{{ stripLastDot $OutputType $Namespace }}, error) {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}_Done_" + client.ID

	result, err := client.Service.Bus.Connection.RequestWithContext(client.Context, functionName, nil)

	if err != nil {
		return nil, errors.New(functionName + ":" + err.Error())
	}

	if result.Data[0] == 0 {
		// 0 means no error
		p := &{{ stripLastDot $OutputType $Namespace }}{}
		err = proto.Unmarshal(result.Data[1:], p)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		var pErr toldata.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return nil, errors.New(pErr.ErrorMessage)
		} else {
			return nil, err
		}
	}
}

func (impl *{{ $ServiceName }}_{{ .Name }}ToldataServerImpl) Subscribe(service *{{ $ServiceName }}ToldataServer, id string) error {
	bus := service.Bus
	var sub *nats.Subscription
	var subscriptions []*nats.Subscription
	var err error

	{{ if .ClientStreaming }}
	sub, err = bus.Connection.QueueSubscribe("{{ $Namespace}}/{{ $ServiceName }}/{{ .Name }}_Send_"+id, "{{ $Namespace}}/{{ $ServiceName }}", func(m *nats.Msg) {
		var input {{ stripLastDot $InputType $Namespace }}
		err := proto.Unmarshal(m.Data, &input)
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}

		err = impl.OnData(&input)

		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		} else {
			zero := []byte{0}
			bus.Connection.Publish(m.Reply, zero)
		}

	})

	subscriptions = append(subscriptions, sub)

	sub, err = bus.Connection.QueueSubscribe("{{ $Namespace}}/{{ $ServiceName }}/{{ .Name }}_Done_"+id, "{{ $Namespace}}/{{ $ServiceName }}", func(m *nats.Msg) {

		defer impl.Exit()
		impl.TriggerEOF()
		result, err := impl.GetResponse()

		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}
		raw, err := proto.Marshal(result)
		if err != nil {
			bus.HandleError(m.Reply, err)
		} else {
			zero := []byte{0}
			bus.Connection.Publish(m.Reply, append(zero, raw...))
		}

	})

	subscriptions = append(subscriptions, sub)

	{{ end }}

	{{ if .ServerStreaming }}
	sub, err = bus.Connection.QueueSubscribe("{{ $Namespace}}/{{ $ServiceName }}/{{ .Name }}_Receive_"+id, "{{ $Namespace}}/{{ $ServiceName }}", func(m *nats.Msg) {
		var input {{ stripLastDot $InputType $Namespace }}
		err := proto.Unmarshal(m.Data, &input)
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}

		response, err := impl.GetResponse()
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}

		raw, err := proto.Marshal(response)
		if err != nil {
			bus.HandleError(m.Reply, err)
		} else {
			zero := []byte{0}
			bus.Connection.Publish(m.Reply, append(zero, raw...))
		}

	})

	subscriptions = append(subscriptions, sub)
	{{ end }}


	impl.OnExit(func() {
			for i := range subscriptions {
				subscriptions[i].Unsubscribe()
			}
	})

	return err
}




{{ if .ServerStreaming }}
func (service *{{ $ServiceName }}ToldataClient) {{ .Name }}(ctx context.Context, req *{{ stripLastDot $InputType $Namespace }}) (*{{ $ServiceName }}ToldataClient_{{ .Name }}, error) {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}"
	if req == nil {
		return nil, errors.New("empty-request")
	}
	reqRaw, err := proto.Marshal(req)	
	result, err := service.Bus.Connection.RequestWithContext(ctx, functionName, reqRaw)
{{ else }}
func (service *{{ $ServiceName }}ToldataClient) {{ .Name }}(ctx context.Context) (*{{ $ServiceName }}ToldataClient_{{ .Name }}, error) {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}"
	
	result, err := service.Bus.Connection.RequestWithContext(ctx, functionName, nil)

{{ end }}
	if err != nil {
		return nil, errors.New(functionName + ":" + err.Error())
	}

	if result.Data[0] == 0 {
		// 0 means no error

		p := &toldata.StreamInfo{}
		err = proto.Unmarshal(result.Data[1:], p)
		if err != nil {
			return nil, err
		}
		return &{{ $ServiceName }}ToldataClient_{{ .Name }}{
			ID:      p.ID,
			Context: ctx,
			Service: service,
		}, nil
	} else {
		var pErr toldata.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return nil, errors.New(pErr.ErrorMessage)
		} else {
			return nil, err
		}
	}
}

{{ else }}

func (service *{{ $ServiceName }}ToldataClient) {{ .Name }}(ctx context.Context, req *{{ stripLastDot $InputType $Namespace }}) (*{{ stripLastDot $OutputType $Namespace }}, error) {
	functionName := "{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}"
	
	if req == nil {
		return nil, errors.New("empty-request")
	}
	reqRaw, err := proto.Marshal(req)

	result, err := service.Bus.Connection.RequestWithContext(ctx, functionName, reqRaw)
	if err != nil {
		return nil, errors.New(functionName + ":" + err.Error())
	}

	if result.Data[0] == 0 {
		// 0 means no error
		p := &{{ stripLastDot $OutputType $Namespace }}{}
		err = proto.Unmarshal(result.Data[1:], p)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		var pErr toldata.ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)
		if err == nil {
			return nil, errors.New(pErr.ErrorMessage)
		} else {
			return nil, err
		}
	}
}

{{ end }}

{{ end }}
{{ end }}

{{ range .Services }}{{ $ServiceName := .Name }}


func (service *{{ $ServiceName }}ToldataServer) Subscribe{{ .Name }}() (<-chan struct{}, error) {
	bus := service.Bus
	
	var err error
	var sub *nats.Subscription
	var subscriptions []*nats.Subscription
	
	done := make(chan struct{})
	
	{{ range .Method }}	

{{ $InputType := .InputType }}
{{ $OutputType := .OutputType }}
	{{ if or .ClientStreaming .ServerStreaming }}
	sub, err = bus.Connection.QueueSubscribe("{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}", "{{ $Namespace}}/{{ $ServiceName }}", func(m *nats.Msg) {
		stream := Create{{ $ServiceName }}_{{ .Name }}ToldataServerImpl(bus.Context)

		

		stream.Subscribe(service, m.Reply)

		raw, err := proto.Marshal(&toldata.StreamInfo{
			ID: m.Reply,
		})
		if err != nil {
			bus.HandleError(m.Reply, err)
		} else {
			zero := []byte{0}
			bus.Connection.Publish(m.Reply, append(zero, raw...))
		}
		{{ if .ServerStreaming }}
		var input {{ stripLastDot $InputType $Namespace }}
		err = proto.Unmarshal(m.Data, &input)
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}
		err = service.Service.{{ .Name }}(&input, stream)
		if err != nil {
			stream.Error(err)
			bus.HandleError(m.Reply, err)
			return
		} else {
			zero := []byte{0}
			bus.Connection.Publish(m.Reply, zero)	
		}
		stream.TriggerEOF()
		{{ else }}
		service.Service.{{ .Name }}(stream)
		{{ end }}
	})

	subscriptions = append(subscriptions, sub)

	{{ else }}
	sub, err = bus.Connection.QueueSubscribe("{{ $Namespace }}/{{ $ServiceName }}/{{ .Name }}", "{{ $Namespace}}/{{ $ServiceName }}", func(m *nats.Msg) {
		var input {{ stripLastDot $InputType $Namespace }}
		err := proto.Unmarshal(m.Data, &input)
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}
		result, err := service.Service.{{ .Name }}(bus.Context, &input)

		if m.Reply != ""  {
			if err != nil {
				bus.HandleError(m.Reply, err)
			} else {
				raw, err := proto.Marshal(result)
				if err != nil {
					bus.HandleError(m.Reply, err)
				} else {
					zero := []byte{0}
					bus.Connection.Publish(m.Reply, append(zero, raw...))
				}
			}
		}

	})

	subscriptions = append(subscriptions, sub)
	{{ end }}



	{{ end }}


	sub, err = bus.Connection.QueueSubscribe("{{ $Namespace }}/{{ $ServiceName }}/ToldataHealthCheck", "{{ $Namespace}}/{{ $ServiceName }}", func(m *nats.Msg) {
		var input toldata.Empty
		err := proto.Unmarshal(m.Data, &input)
		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}
		result, err := service.Service.ToldataHealthCheck(bus.Context, &input)

		if m.Reply != ""  {
			if err != nil {
				bus.HandleError(m.Reply, err)
			} else {
				raw, err := proto.Marshal(result)
				if err != nil {
					bus.HandleError(m.Reply, err)
				} else {
					zero := []byte{0}
					bus.Connection.Publish(m.Reply, append(zero, raw...))
				}
			}
		}

	})

	subscriptions = append(subscriptions, sub)


	go func() {
		defer close(done)

		select {
		case <-bus.Context.Done():
			for i := range subscriptions {
				subscriptions[i].Unsubscribe()
			}
		}
	}()

	return done, err
}



{{ end }}


`
)
