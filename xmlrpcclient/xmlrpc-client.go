package xmlrpcclient

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/AlexStocks/gorilla-xmlrpc/xml"
	"github.com/AlexStocks/supervisord/types"
)

type XmlRPCClient struct {
	serverurl string
	user      string
	password  string
	timeout   time.Duration
	verbose   bool
}

type VersionReply struct {
	Value string
}

type StartStopReply struct {
	Value bool
}

type RestartReply struct {
	StopValue  bool
	StartValue bool
}

type ShutdownReply StartStopReply

type AllProcessInfoReply struct {
	Value []types.ProcessInfo
}

type ListMethodsReply struct {
	Methods []string
}

var emptyReader io.ReadCloser

func init() {
	var buf bytes.Buffer
	emptyReader = ioutil.NopCloser(&buf)
}

func NewXmlRPCClient(serverurl string, verbose bool) *XmlRPCClient {
	return &XmlRPCClient{serverurl: serverurl, timeout: 0, verbose: verbose}
}

func (r *XmlRPCClient) SetUser(user string) {
	r.user = user
}

func (r *XmlRPCClient) SetPassword(password string) {
	r.password = password
}

func (r *XmlRPCClient) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

func (r *XmlRPCClient) Url() string {
	return fmt.Sprintf("%s/RPC2", r.serverurl)
}

func (r *XmlRPCClient) createHttpRequest(method string, url string, data interface{}) (*http.Request, error) {
	buf, _ := xml.EncodeClientRequest(method, data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		if r.verbose {
			fmt.Println("Fail to create request:", err)
		}
		return nil, err
	}

	if len(r.user) > 0 && len(r.password) > 0 {
		req.SetBasicAuth(r.user, r.password)
	}

	req.Header.Set("Content-Type", "text/xml")

	return req, nil
}

func (r *XmlRPCClient) processResponse(resp *http.Response, processBody func(io.ReadCloser, error)) {
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		if r.verbose {
			fmt.Println("Bad Response:", resp.Status)
		}
		processBody(emptyReader, fmt.Errorf("Bad response with status %d", resp.Status))
	} else {
		processBody(resp.Body, nil)
	}
}

func (r *XmlRPCClient) postInetHttp(method string, url string, data interface{}, processBody func(io.ReadCloser, error)) {
	req, err := r.createHttpRequest(method, url, data)
	if err != nil {
		return
	}

	if r.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Fail to send request to supervisord:", err)
		os.Exit(2)
		return
	}
	r.processResponse(resp, processBody)
}

func (r *XmlRPCClient) postUnixHttp(method string, path string, data interface{}, processBody func(io.ReadCloser, error)) {
	var conn net.Conn
	var err error
	if r.timeout > 0 {
		conn, err = net.DialTimeout("unix", path, r.timeout)
	} else {
		conn, err = net.Dial("unix", path)
	}
	if err != nil {
		if r.verbose {
			fmt.Printf("Fail to connect unix socket path: %s\n", r.serverurl)
		}
		return
	}
	defer conn.Close()

	if r.timeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(r.timeout)); err != nil {
			return
		}
	}
	req, err := r.createHttpRequest(method, "/RPC2", data)

	if err != nil {
		return
	}
	err = req.Write(conn)
	if err != nil {
		if r.verbose {
			fmt.Printf("Fail to write to unix socket %s\n", r.serverurl)
		}
		return
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		if r.verbose {
			fmt.Printf("Fail to read response %s\n", err)
		}
		return
	}
	r.processResponse(resp, processBody)
}

func (r *XmlRPCClient) post(method string, data interface{}, processBody func(io.ReadCloser, error)) {
	url, err := url.Parse(r.serverurl)
	if err != nil {
		fmt.Printf("Malform url:%s\n", url)
		return
	}
	if url.Scheme == "http" || url.Scheme == "https" {
		r.postInetHttp(method, r.Url(), data, processBody)
	} else if url.Scheme == "unix" {
		r.postUnixHttp(method, url.Path, data, processBody)
	} else {
		fmt.Printf("Unsupported URL scheme:%s\n", url.Scheme)
	}
}

func (r *XmlRPCClient) GetVersion() (reply VersionReply, err error) {
	ins := struct{}{}
	r.post("supervisor.getVersion", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, reply)
		}
	})
	return
}

func (r *XmlRPCClient) GetAllProcessInfo() (reply AllProcessInfoReply, err error) {
	ins := struct{}{}
	r.post("supervisor.getAllProcessInfo", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})

	return
}

func (r *XmlRPCClient) GetAllProcsProcessInfo() (reply AllProcessInfoReply, err error) {
	ins := struct{}{}
	r.post("supervisor.getAllProcsProcessInfo", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})

	return
}

func (r *XmlRPCClient) GetAllInfomapProcessInfo() (reply AllProcessInfoReply, err error) {
	ins := struct{}{}
	r.post("supervisor.getAllInfomapProcessInfo", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})

	return
}

func (r *XmlRPCClient) GetPrestartProcessInfo() (reply AllProcessInfoReply, err error) {
	ins := struct{}{}
	r.post("supervisor.getPrestartProcessInfo", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})

	return
}

func (r *XmlRPCClient) ChangeProcessState(change string, processName string) (reply interface{}, err error) {
	if !(change == "start" || change == "stop" || change == "restart" || change == "remove") {
		err = fmt.Errorf("Incorrect required state")
		return
	}
	ins := struct{ Value string }{processName}
	method := fmt.Sprintf("supervisor.%sProcess", change)
	if change == "remove" {
		method = fmt.Sprintf("supervisor.%sProcessGroup", change)
	}
	r.post(method, &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			if change != "restart" {
				var startStopReply StartStopReply
				err = xml.DecodeClientResponse(body, &startStopReply)
				reply = startStopReply
			} else {
				var restartReply RestartReply
				err = xml.DecodeClientResponse(body, &restartReply)
				reply = restartReply
			}
		}
	})

	return
}

func (r *XmlRPCClient) ChangeAllProcessState(change string) (reply AllProcessInfoReply, err error) {
	// if !(change == "start" || change == "stop" || change == "restart" || change == "remove") {
	if !(change == "start" || change == "stop" || change == "restart") {
		err = fmt.Errorf("Incorrect required state")
		return
	}
	ins := struct{ Wait bool }{true}
	r.post(fmt.Sprintf("supervisor.%sAllProcesses", change), &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})
	return
}

func (r *XmlRPCClient) Shutdown() (reply ShutdownReply, err error) {
	ins := struct{}{}
	r.post("supervisor.shutdown", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}

	})

	return
}

func (r *XmlRPCClient) ReloadConfig() (reply types.ReloadConfigResult, err error) {
	ins := struct{}{}

	xmlProcMgr := NewXmlProcessorManager()
	reply.AddedGroup = make([]string, 0)
	reply.ChangedGroup = make([]string, 0)
	reply.RemovedGroup = make([]string, 0)
	i := 0
	has_value := false
	xmlProcMgr.AddNonLeafProcessor("methodResponse/params/param/value/array/data", func() {
		i++
		if has_value {
			has_value = false
		}
		// else {
		// 	i++
		// }
	})
	xmlProcMgr.AddLeafProcessor("methodResponse/params/param/value/array/data/value/string", func(value string) {
		has_value = true
		switch i {
		case 0:
			reply.AddedGroup = append(reply.AddedGroup, value)
		case 1:
			reply.ChangedGroup = append(reply.ChangedGroup, value)
		case 2:
			reply.RemovedGroup = append(reply.RemovedGroup, value)
		}
	})
	r.post("supervisor.reloadConfig", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			xmlProcMgr.ProcessXml(body)
		}
	})
	return
}

func (r *XmlRPCClient) Update(process string) (reply types.UpdateResult, err error) {
	ins := struct{ Process string }{Process: process}

	xmlProcMgr := NewXmlProcessorManager()
	reply.AddedGroup = make([]string, 0)
	reply.ChangedGroup = make([]string, 0)
	reply.RemovedGroup = make([]string, 0)
	i := 0
	has_value := false
	xmlProcMgr.AddNonLeafProcessor("methodResponse/params/param/value/array/data", func() {
		i++
		if has_value {
			has_value = false
		} // else {
		// 	i++
		// }
	})
	xmlProcMgr.AddLeafProcessor("methodResponse/params/param/value/array/data/value/string", func(value string) {
		has_value = true
		switch i {
		case 0:
			reply.AddedGroup = append(reply.AddedGroup, value)
		case 1:
			reply.ChangedGroup = append(reply.ChangedGroup, value)
		case 2:
			reply.RemovedGroup = append(reply.RemovedGroup, value)
		}
	})
	xmlProcMgr.AddLeafProcessor("methodResponse/fault/value/struct/member/value/string", func(value string) {
		err = errors.New(value)
	})
	r.post("supervisor.update", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			xmlProcMgr.ProcessXml(body)
		} else {
			fmt.Println("ProcessXmlRpc ERROR", err)
		}
	})
	return
}

func (r *XmlRPCClient) UpdateAll() (reply types.UpdateResult, err error) {
	ins := struct{}{}

	xmlProcMgr := NewXmlProcessorManager()
	reply.AddedGroup = make([]string, 0)
	reply.ChangedGroup = make([]string, 0)
	reply.RemovedGroup = make([]string, 0)
	i := 0
	has_value := false
	xmlProcMgr.AddNonLeafProcessor("methodResponse/params/param/value/array", func() {
		i++
		if has_value {
			has_value = false
		} // else {
		// 	i++
		// }
	})
	xmlProcMgr.AddLeafProcessor("methodResponse/params/param/value/array/data/value/string", func(value string) {
		has_value = true
		switch i {
		case 0:
			reply.AddedGroup = append(reply.AddedGroup, value)
		case 1:
			reply.ChangedGroup = append(reply.ChangedGroup, value)
		case 2:
			reply.RemovedGroup = append(reply.RemovedGroup, value)
		}
	})
	r.post("supervisor.updateAll", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			xmlProcMgr.ProcessXml(body)
		}
	})
	return
}

func (r *XmlRPCClient) SignalProcess(signal string, name string) (reply types.BooleanReply, err error) {
	ins := types.ProcessSignal{Name: name, Signal: signal}
	r.post("supervisor.signalProcess", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})
	return
}

func (r *XmlRPCClient) SignalAll(signal string) (reply AllProcessInfoReply, err error) {
	ins := struct{ Signal string }{signal}
	// r.post("supervisor.signalProcess", &ins, func(body io.ReadCloser, procError error) {
	r.post("supervisor.signalAllProcesses", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &reply)
		}
	})

	return
}

func (r *XmlRPCClient) GetProcessInfo(process string) (reply types.ProcessInfo, err error) {
	ins := struct{ Name string }{process}
	result := struct{ Reply types.ProcessInfo }{}
	r.post("supervisor.getProcessInfo", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &result)
			if err == nil {
				reply = result.Reply
			} else if r.verbose {
				fmt.Printf("Fail to decode to types.ProcessInfo\n")
			}
		}
	})

	return
}

func (r *XmlRPCClient) ListMethods() (reply ListMethodsReply, err error) {
	ins := struct{}{}
	result := ListMethodsReply{}
	result.Methods = make([]string, 0)
	r.post("supervisor.listMethods", &ins, func(body io.ReadCloser, procError error) {
		err = procError
		if err == nil {
			err = xml.DecodeClientResponse(body, &result)
			if err == nil {
				reply = result
			} else if r.verbose {
				fmt.Printf("Fail to decode to ListMethodsReply\n")
			}
		}
	})

	return
}
