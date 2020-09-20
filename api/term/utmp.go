package term

/*
map[type:UserProcess user:root exit:{0 0} session:0 time:Thu, 06 Feb 2020 07:20:50 -0500 address:192.168.75.1 pid:55555 device:pts/2 id:ts/2 host:192.168.75.1]
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Empty        = 0x0
	RunLevel     = 0x1
	BootTime     = 0x2
	NewTime      = 0x3
	OldTime      = 0x4
	InitProcess  = 0x5
	LoginProcess = 0x6
	UserProcess  = 0x7
	DeadProcess  = 0x8
	Accounting   = 0x9
)

const (
	LineSize = 32
	NameSize = 32
	HostSize = 256
)

// utmp structures
// see man utmp
type ExitStatus struct {
	Termination int16
	Exit        int16
}

type TimeVal struct {
	Sec  int32
	Usec int32
}

func (t TimeVal) humanTime() string {
	ts := time.Unix(int64(t.Sec), int64(t.Usec))
	return string(ts.Format(time.RFC1123Z))
}

type Utmp struct {
	Type int16
	// alignment
	_       [2]byte
	Pid     int32
	Device  [LineSize]byte
	Id      [4]byte
	User    [NameSize]byte
	Host    [HostSize]byte
	Exit    ExitStatus
	Session int32
	Time    TimeVal
	Addr    [4]int32 // Internet address of remote host; IPv4 address uses just Addr[0]
	// AddrV6  [16]byte
	// Reserved member
	Reserved [20]byte
}

func newUtmp() *Utmp {
	return &Utmp{
		Type:   UserProcess,
		Pid:    10001,
		Device: toByte32("pts/6"),
		Id:     toByte4(idSuffix("pts/6")),
		User:   toByte32("root2"),
		Host:   toByte256("127.0.0.1"),
		Time: TimeVal{
			Sec:  int32(time.Now().Unix()),
			Usec: int32(time.Now().Second()),
		},
		Addr: [4]int32{addrNum("192.168.75.1")},
	}
}

func idSuffix(s string) string {
	if len(s) > 4 {
		return s[len(s)-4:]
	}
	return s
}

func addrNum(s string) int32 {
	var num int32
	strs := strings.Split(s, ".")
	for i := len(strs) - 1; i > -1; i-- {
		n, _ := strconv.ParseInt(strs[i], 10, 32)
		fmt.Println(n)
		if num == 0 {
			num = int32(n)
		} else {
			num = num<<8 + int32(n)
		}
	}
	return num
}

func toByte4(s string) [4]byte {
	var a [4]byte
	for i, b := range s {
		a[i] = byte(b)
	}
	return a
}
func toByte32(s string) [32]byte {
	var a [32]byte
	for i, b := range s {
		a[i] = byte(b)
	}
	return a
}

func toByte256(s string) [256]byte {
	var a [256]byte
	for i, b := range s {
		a[i] = byte(b)
	}
	return a
}
func init() {
	// readUtmp()
	// fmt.Printf("%#v\n", newUtmp())
}
func readUtmp() {
	file, err := os.Open("/var/run/utmp")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	for {
		var nu Utmp
		err := binary.Read(file, binary.LittleEndian, &nu)
		if err == io.EOF {
			break
		}
		if nu.Type == UserProcess {
			fmt.Printf("%#v\n", nu)
			hhhh(&nu)
		}
	}
}

func UtmpWrite(utmp *Utmp) error {
	source := "/var/run/utmp"
	file, err := os.OpenFile(source, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return err
	}
	return binary.Write(file, binary.LittleEndian, utmp)
}
func humanType(u int16) string {
	switch u {
	case Empty:
		return "Empty"
	case RunLevel:
		return "RunLevel"
	case BootTime:
		return "BootTime"
	case NewTime:
		return "NewTime"
	case OldTime:
		return "OldTime"
	case InitProcess:
		return "InitProcess"
	case LoginProcess:
		return "LoginProcess"
	case UserProcess:
		return "UserProcess"
	case DeadProcess:
		return "DeadProcess"
	case Accounting:
		return "Accounting"
	default:
		return ""
	}
}
func AddrToString(a [4]int32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(a[0]), byte(a[0]>>8), byte(a[0]>>16), byte(a[0]>>24))
}
func hhhh(u *Utmp) {
	utmp := map[string]interface{}{}
	utmp["type"] = humanType(u.Type)
	utmp["pid"] = u.Pid
	utmp["device"] = string(bytes.Trim(u.Device[:], "\u0000"))
	utmp["id"] = string(bytes.Trim(u.Id[:], "\u0000"))
	utmp["user"] = string(bytes.Trim(u.User[:], "\u0000"))
	utmp["host"] = string(bytes.Trim(u.Host[:], "\u0000"))
	utmp["exit"] = u.Exit
	utmp["session"] = u.Session
	utmp["time"] = u.Time.humanTime()
	utmp["address"] = AddrToString(u.Addr)
	fmt.Println(utmp)
}
