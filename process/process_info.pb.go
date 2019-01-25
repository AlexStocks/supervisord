// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: process_info.proto

package process

import (
	fmt "fmt"
	"os"
	"path"
	"syscall"
	"time"

	math "math"
	reflect "reflect"
	strings "strings"

	"github.com/AlexStocks/goext/os/process"
	"github.com/AlexStocks/supervisord/config"
	"github.com/AlexStocks/supervisord/signals"
	"github.com/AlexStocks/supervisord/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_sortkeys "github.com/gogo/protobuf/sortkeys"
	_ "github.com/mwitkow/go-proto-validators"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	io "io"
	"io/ioutil"

	jerrors "github.com/juju/errors"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type ProcessInfo struct {
	//  @inject_tag: yaml:"start_time"
	StartTime uint64 `protobuf:"varint,1,opt,name=StartTime" json:"StartTime" yaml:"start_time"`
	//  @inject_tag: yaml:"pid"
	PID uint64 `protobuf:"varint,2,opt,name=PID" json:"PID" yaml:"pid"`
	//  @inject_tag: yaml:"program"
	Program string `protobuf:"bytes,3,opt,name=Program" json:"Program" yaml:"program"`
	config  *config.ConfigEntry
}

func (p *ProcessInfo) TypeProcessInfo() types.ProcessInfo {
	state := ProcessState(RUNNING)
	if _, err := gxprocess.FindProcess(int(p.PID)); err != nil {
		state = ProcessState(STOPPED)
	}
	info := types.ProcessInfo{
		Name: p.Program,
		// Group:          p.GetGroup(),
		// Description:    p.GetDescription(),
		Start: int(p.StartTime) / 1e9,
		// Stop:           int(p.GetStopTime().Unix()),
		Now:       int(time.Now().Unix()),
		State:     int(state),
		Statename: state.String(),
		Spawnerr:  "",
		// Exitstatus:     0,
		Logfile:        getStdoutLogfile(p.config),
		Stdout_logfile: getStdoutLogfile(p.config),
		Stderr_logfile: getStderrLogfile(p.config),
		Pid:            int(p.PID),
	}

	startTime := time.Unix(int64(p.StartTime/1e9), int64(p.StartTime%1e9))
	seconds := int(time.Now().Sub(startTime).Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	if days > 0 {
		info.Description = fmt.Sprintf("pid %d, uptime %d days, %d:%02d:%02d", info.Pid, days, hours%24, minutes%60, seconds%60)
	} else {
		info.Description = fmt.Sprintf("pid %d, uptime %d:%02d:%02d", info.Pid, hours%24, minutes%60, seconds%60)
	}

	return info
}

func (p *ProcessInfo) ConfigEntry() *config.ConfigEntry {
	return p.config
}

//send signal to process to stop it
func (p *ProcessInfo) Stop(wait bool) {
	log.WithFields(log.Fields{"program": p.Program}).Info("stop the program")
	sigs := strings.Fields(p.config.GetString("stopsignal", ""))
	waitsecs := time.Duration(p.config.GetInt("stopwaitsecs", 10)) * time.Second
	stopasgroup := p.config.GetBool("stopasgroup", false)
	killasgroup := p.config.GetBool("killasgroup", stopasgroup)
	if stopasgroup && !killasgroup {
		log.WithFields(log.Fields{"program": p.Program}).Error("Cannot set stopasgroup=true and killasgroup=false")
	}

	go func() {
		stopped := false
		for i := 0; i < len(sigs) && !stopped; i++ {
			// send signal to process
			sig, err := signals.ToSignal(sigs[i])
			if err != nil {
				continue
			}
			log.WithFields(log.Fields{"program": p.Program, "signal": sigs[i], "pid": p.PID}).Info("send stop signal to program")
			signals.KillPid(int(p.PID), sig, stopasgroup)
			endTime := time.Now().Add(waitsecs)
			//wait at most "stopwaitsecs" seconds for one signal
			for endTime.After(time.Now()) {
				//if it already exits
				if _, err := gxprocess.FindProcess(int(p.PID)); err != nil {
					stopped = true
					break
				}

				time.Sleep(1 * time.Second)
			}
		}
		if !stopped {
			log.WithFields(log.Fields{"program": p.Program, "signal": "KILL", "pid": p.PID}).Info("force to kill the program")
			signals.KillPid(int(p.PID), syscall.SIGKILL, killasgroup)
		}
	}()
	if wait {
		for {
			if _, err := gxprocess.FindProcess(int(p.PID)); err != nil {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (m *ProcessInfo) Reset()      { *m = ProcessInfo{} }
func (*ProcessInfo) ProtoMessage() {}
func (*ProcessInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_process_info_ed4b98fde5854bca, []int{0}
}
func (m *ProcessInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProcessInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProcessInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *ProcessInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProcessInfo.Merge(dst, src)
}
func (m *ProcessInfo) XXX_Size() int {
	return m.Size()
}
func (m *ProcessInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ProcessInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ProcessInfo proto.InternalMessageInfo

type ProcessInfoMap struct {
	//  @inject_tag: yaml:"version"
	Version uint64 `protobuf:"varint,1,opt,name=Version" json:"Version" yaml:"version"`
	//  @inject_tag: yaml:"info_map"
	InfoMap map[string]ProcessInfo `protobuf:"bytes,2,rep,name=InfoMap" json:"InfoMap" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value" yaml:"info_map"`
}

func NewProcessInfoMap() *ProcessInfoMap {
	return &ProcessInfoMap{
		InfoMap: make(map[string]ProcessInfo),
	}
}

func (m *ProcessInfoMap) Load(file string) error {
	configFile, err := ioutil.ReadFile(file)
	if err != nil {
		return jerrors.Trace(err)
	}

	err = yaml.Unmarshal(configFile, m)
	if err != nil {
		return jerrors.Trace(err)
	}

	return nil
}

func (m *ProcessInfoMap) Store(file string) error {
	if err := m.Validate(); err != nil {
		return err
	}

	// valid info map
	infoMap := NewProcessInfoMap()
	infoMap.Version = m.Version
	for _, info := range m.InfoMap {
		if _, err := gxprocess.FindProcess(int(info.PID)); err == nil {
			infoMap.AddProcessInfo(info)
		}
	}

	var fileStream []byte
	fileStream, err := yaml.Marshal(infoMap)
	if err != nil {
		return jerrors.Trace(err)
	}

	basePath := path.Dir(file)
	if err = os.MkdirAll(basePath, 0766); err != nil &&
		!strings.Contains(err.Error(), "file exists") {
		return jerrors.Trace(err)
	}
	os.Remove(file)

	err = ioutil.WriteFile(file, fileStream, 0766)
	if err != nil {
		return jerrors.Trace(err)
	}

	return nil
}

func (m *ProcessInfoMap) AddProcessInfo(info ProcessInfo) {
	m.InfoMap[info.Program] = info
	m.Version = uint64(time.Now().UnixNano())
}

func (m *ProcessInfoMap) RemoveProcessInfo(program string) ProcessInfo {
	info, ok := m.InfoMap[program]
	if ok {
		delete(m.InfoMap, program)
	}
	m.Version = uint64(time.Now().UnixNano())
	return info
}

func (m *ProcessInfoMap) GetProcessInfo(program string) (ProcessInfo, bool) {
	info, ok := m.InfoMap[program]
	return info, ok
}

func (m *ProcessInfoMap) Reset()      { *m = ProcessInfoMap{} }
func (*ProcessInfoMap) ProtoMessage() {}
func (*ProcessInfoMap) Descriptor() ([]byte, []int) {
	return fileDescriptor_process_info_ed4b98fde5854bca, []int{1}
}
func (m *ProcessInfoMap) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProcessInfoMap) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProcessInfoMap.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *ProcessInfoMap) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProcessInfoMap.Merge(dst, src)
}
func (m *ProcessInfoMap) XXX_Size() int {
	return m.Size()
}
func (m *ProcessInfoMap) XXX_DiscardUnknown() {
	xxx_messageInfo_ProcessInfoMap.DiscardUnknown(m)
}

var xxx_messageInfo_ProcessInfoMap proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ProcessInfo)(nil), "process.ProcessInfo")
	proto.RegisterType((*ProcessInfoMap)(nil), "process.ProcessInfoMap")
	proto.RegisterMapType((map[string]ProcessInfo)(nil), "process.ProcessInfoMap.InfoMapEntry")
}
func (this *ProcessInfo) VerboseEqual(that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*ProcessInfo)
	if !ok {
		that2, ok := that.(ProcessInfo)
		if ok {
			that1 = &that2
		} else {
			return fmt.Errorf("that is not of type *ProcessInfo")
		}
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *ProcessInfo but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *ProcessInfo but is not nil && this == nil")
	}
	if this.StartTime != that1.StartTime {
		return fmt.Errorf("StartTime this(%v) Not Equal that(%v)", this.StartTime, that1.StartTime)
	}
	if this.PID != that1.PID {
		return fmt.Errorf("PID this(%v) Not Equal that(%v)", this.PID, that1.PID)
	}
	if this.Program != that1.Program {
		return fmt.Errorf("Program this(%v) Not Equal that(%v)", this.Program, that1.Program)
	}
	return nil
}
func (this *ProcessInfo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ProcessInfo)
	if !ok {
		that2, ok := that.(ProcessInfo)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.StartTime != that1.StartTime {
		return false
	}
	if this.PID != that1.PID {
		return false
	}
	if this.Program != that1.Program {
		return false
	}
	return true
}
func (this *ProcessInfoMap) VerboseEqual(that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*ProcessInfoMap)
	if !ok {
		that2, ok := that.(ProcessInfoMap)
		if ok {
			that1 = &that2
		} else {
			return fmt.Errorf("that is not of type *ProcessInfoMap")
		}
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *ProcessInfoMap but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *ProcessInfoMap but is not nil && this == nil")
	}
	if this.Version != that1.Version {
		return fmt.Errorf("Version this(%v) Not Equal that(%v)", this.Version, that1.Version)
	}
	if len(this.InfoMap) != len(that1.InfoMap) {
		return fmt.Errorf("InfoMap this(%v) Not Equal that(%v)", len(this.InfoMap), len(that1.InfoMap))
	}
	for i := range this.InfoMap {
		a := this.InfoMap[i]
		b := that1.InfoMap[i]
		if !(&a).Equal(&b) {
			return fmt.Errorf("InfoMap this[%v](%v) Not Equal that[%v](%v)", i, this.InfoMap[i], i, that1.InfoMap[i])
		}
	}
	return nil
}
func (this *ProcessInfoMap) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ProcessInfoMap)
	if !ok {
		that2, ok := that.(ProcessInfoMap)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Version != that1.Version {
		return false
	}
	if len(this.InfoMap) != len(that1.InfoMap) {
		return false
	}
	for i := range this.InfoMap {
		a := this.InfoMap[i]
		b := that1.InfoMap[i]
		if !(&a).Equal(&b) {
			return false
		}
	}
	return true
}
func (this *ProcessInfo) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&process.ProcessInfo{")
	s = append(s, "StartTime: "+fmt.Sprintf("%#v", this.StartTime)+",\n")
	s = append(s, "PID: "+fmt.Sprintf("%#v", this.PID)+",\n")
	s = append(s, "Program: "+fmt.Sprintf("%#v", this.Program)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *ProcessInfoMap) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&process.ProcessInfoMap{")
	s = append(s, "Version: "+fmt.Sprintf("%#v", this.Version)+",\n")
	keysForInfoMap := make([]string, 0, len(this.InfoMap))
	for k, _ := range this.InfoMap {
		keysForInfoMap = append(keysForInfoMap, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForInfoMap)
	mapStringForInfoMap := "map[string]ProcessInfo{"
	for _, k := range keysForInfoMap {
		mapStringForInfoMap += fmt.Sprintf("%#v: %#v,", k, this.InfoMap[k])
	}
	mapStringForInfoMap += "}"
	if this.InfoMap != nil {
		s = append(s, "InfoMap: "+mapStringForInfoMap+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringProcessInfo(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *ProcessInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProcessInfo) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintProcessInfo(dAtA, i, uint64(m.StartTime))
	dAtA[i] = 0x10
	i++
	i = encodeVarintProcessInfo(dAtA, i, uint64(m.PID))
	dAtA[i] = 0x1a
	i++
	i = encodeVarintProcessInfo(dAtA, i, uint64(len(m.Program)))
	i += copy(dAtA[i:], m.Program)
	return i, nil
}

func (m *ProcessInfoMap) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProcessInfoMap) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintProcessInfo(dAtA, i, uint64(m.Version))
	if len(m.InfoMap) > 0 {
		for k, _ := range m.InfoMap {
			dAtA[i] = 0x12
			i++
			v := m.InfoMap[k]
			msgSize := 0
			if (&v) != nil {
				msgSize = (&v).Size()
				msgSize += 1 + sovProcessInfo(uint64(msgSize))
			}
			mapSize := 1 + len(k) + sovProcessInfo(uint64(len(k))) + msgSize
			i = encodeVarintProcessInfo(dAtA, i, uint64(mapSize))
			dAtA[i] = 0xa
			i++
			i = encodeVarintProcessInfo(dAtA, i, uint64(len(k)))
			i += copy(dAtA[i:], k)
			dAtA[i] = 0x12
			i++
			i = encodeVarintProcessInfo(dAtA, i, uint64((&v).Size()))
			n1, err := (&v).MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n1
		}
	}
	return i, nil
}

func encodeVarintProcessInfo(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ProcessInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovProcessInfo(uint64(m.StartTime))
	n += 1 + sovProcessInfo(uint64(m.PID))
	l = len(m.Program)
	n += 1 + l + sovProcessInfo(uint64(l))
	return n
}

func (m *ProcessInfoMap) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovProcessInfo(uint64(m.Version))
	if len(m.InfoMap) > 0 {
		for k, v := range m.InfoMap {
			_ = k
			_ = v
			l = v.Size()
			mapEntrySize := 1 + len(k) + sovProcessInfo(uint64(len(k))) + 1 + l + sovProcessInfo(uint64(l))
			n += mapEntrySize + 1 + sovProcessInfo(uint64(mapEntrySize))
		}
	}
	return n
}

func sovProcessInfo(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozProcessInfo(x uint64) (n int) {
	return sovProcessInfo(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *ProcessInfo) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProcessInfo{`,
		`StartTime:` + fmt.Sprintf("%v", this.StartTime) + `,`,
		`PID:` + fmt.Sprintf("%v", this.PID) + `,`,
		`Program:` + fmt.Sprintf("%v", this.Program) + `,`,
		`}`,
	}, "")
	return s
}
func (this *ProcessInfoMap) String() string {
	if this == nil {
		return "nil"
	}
	keysForInfoMap := make([]string, 0, len(this.InfoMap))
	for k, _ := range this.InfoMap {
		keysForInfoMap = append(keysForInfoMap, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForInfoMap)
	mapStringForInfoMap := "map[string]ProcessInfo{"
	for _, k := range keysForInfoMap {
		mapStringForInfoMap += fmt.Sprintf("%v: %v,", k, this.InfoMap[k])
	}
	mapStringForInfoMap += "}"
	s := strings.Join([]string{`&ProcessInfoMap{`,
		`Version:` + fmt.Sprintf("%v", this.Version) + `,`,
		`InfoMap:` + mapStringForInfoMap + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringProcessInfo(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *ProcessInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProcessInfo
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProcessInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProcessInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartTime", wireType)
			}
			m.StartTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StartTime |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PID", wireType)
			}
			m.PID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PID |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Program", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProcessInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Program = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProcessInfo(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProcessInfo
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ProcessInfoMap) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProcessInfo
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProcessInfoMap: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProcessInfoMap: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InfoMap", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthProcessInfo
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.InfoMap == nil {
				m.InfoMap = make(map[string]ProcessInfo)
			}
			var mapkey string
			mapvalue := &ProcessInfo{}
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowProcessInfo
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowProcessInfo
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= (uint64(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthProcessInfo
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowProcessInfo
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= (int(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthProcessInfo
					}
					postmsgIndex := iNdEx + mapmsglen
					if mapmsglen < 0 {
						return ErrInvalidLengthProcessInfo
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &ProcessInfo{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipProcessInfo(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthProcessInfo
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.InfoMap[mapkey] = *mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProcessInfo(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProcessInfo
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipProcessInfo(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProcessInfo
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProcessInfo
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthProcessInfo
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowProcessInfo
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipProcessInfo(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthProcessInfo = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProcessInfo   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("process_info.proto", fileDescriptor_process_info_ed4b98fde5854bca) }

var fileDescriptor_process_info_ed4b98fde5854bca = []byte{
	// 385 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0xcf, 0xca, 0xd3, 0x40,
	0x14, 0xc5, 0xe7, 0x26, 0x95, 0x9a, 0xa9, 0x88, 0x0c, 0x22, 0xa1, 0xc8, 0x34, 0x94, 0x2e, 0x82,
	0xd0, 0x04, 0xba, 0x10, 0x11, 0x57, 0x45, 0x17, 0x15, 0x84, 0x90, 0x48, 0x71, 0x27, 0x69, 0x4d,
	0x63, 0x68, 0x93, 0x09, 0x93, 0xa4, 0xa5, 0x3b, 0x1f, 0xc1, 0x37, 0x70, 0xeb, 0xa3, 0x74, 0xd9,
	0x65, 0x71, 0xa1, 0xcd, 0x74, 0xe3, 0xb2, 0x8f, 0x20, 0xf9, 0x53, 0xd3, 0x8f, 0xaf, 0xab, 0x99,
	0x7b, 0xce, 0xef, 0x9e, 0x7b, 0xb9, 0x98, 0xc4, 0x9c, 0xcd, 0xbd, 0x24, 0xf9, 0x1c, 0x44, 0x0b,
	0x66, 0xc4, 0x9c, 0xa5, 0x8c, 0xb4, 0x6b, 0xad, 0x3b, 0xf4, 0x83, 0xf4, 0x6b, 0x36, 0x33, 0xe6,
	0x2c, 0x34, 0x7d, 0xe6, 0x33, 0xb3, 0xf4, 0x67, 0xd9, 0xa2, 0xac, 0xca, 0xa2, 0xfc, 0x55, 0x7d,
	0xdd, 0x97, 0x57, 0x78, 0xb8, 0x09, 0xd2, 0x25, 0xdb, 0x98, 0x3e, 0x1b, 0x96, 0xe6, 0x70, 0xed,
	0xae, 0x82, 0x2f, 0x6e, 0xca, 0x78, 0x62, 0xfe, 0xff, 0x56, 0x7d, 0xfd, 0x1f, 0x80, 0x3b, 0x56,
	0x35, 0x72, 0x12, 0x2d, 0x18, 0xd1, 0xb1, 0xe2, 0xa4, 0x2e, 0x4f, 0x3f, 0x06, 0xa1, 0xa7, 0x82,
	0x06, 0x7a, 0x6b, 0x8c, 0x77, 0xbf, 0x7b, 0x48, 0xfc, 0xe9, 0x49, 0x4f, 0x90, 0xdd, 0x98, 0xe4,
	0x39, 0x96, 0xad, 0xc9, 0x5b, 0x55, 0xba, 0xc7, 0x14, 0x32, 0x19, 0xe0, 0xb6, 0xc5, 0x99, 0xcf,
	0xdd, 0x50, 0x95, 0x35, 0xd0, 0x95, 0x86, 0xf8, 0x04, 0xf6, 0xc5, 0x22, 0x7d, 0xac, 0x4c, 0x12,
	0x27, 0x76, 0xb2, 0x99, 0xe5, 0xa8, 0x2d, 0x0d, 0xf4, 0x87, 0xe3, 0x56, 0xc1, 0xd9, 0x8d, 0xdc,
	0xff, 0x05, 0xf8, 0xf1, 0xd5, 0x86, 0x1f, 0xdc, 0xb8, 0x08, 0x9f, 0x7a, 0x3c, 0x09, 0x58, 0x74,
	0x63, 0xc5, 0x8b, 0x45, 0xde, 0xe3, 0x76, 0xdd, 0xa0, 0x4a, 0x9a, 0xac, 0x77, 0x46, 0x03, 0xa3,
	0x3e, 0xae, 0x71, 0x37, 0xcf, 0xa8, 0xdf, 0x77, 0x51, 0xca, 0xb7, 0x4d, 0x56, 0x8c, 0xec, 0x4b,
	0x40, 0x77, 0x8a, 0x1f, 0x5d, 0x43, 0xe4, 0x19, 0x96, 0x97, 0xde, 0xb6, 0x9c, 0xae, 0xd4, 0x2b,
	0x17, 0x02, 0x79, 0x81, 0x1f, 0xac, 0xdd, 0x55, 0xe6, 0x95, 0x67, 0xe9, 0x8c, 0x9e, 0xde, 0x9a,
	0x68, 0x57, 0xc8, 0x6b, 0xe9, 0x15, 0x8c, 0xdf, 0xec, 0x72, 0x8a, 0xf6, 0x39, 0x45, 0x87, 0x9c,
	0xa2, 0x63, 0x4e, 0xe1, 0x9c, 0x53, 0xf8, 0x26, 0x28, 0xfc, 0x14, 0x14, 0x76, 0x82, 0xc2, 0x5e,
	0x50, 0x38, 0x0a, 0x0a, 0x7f, 0x05, 0x45, 0x67, 0x41, 0xe1, 0xfb, 0x89, 0xa2, 0xfd, 0x89, 0xa2,
	0xc3, 0x89, 0xa2, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x26, 0xc5, 0xbf, 0x20, 0x41, 0x02, 0x00,
	0x00,
}
