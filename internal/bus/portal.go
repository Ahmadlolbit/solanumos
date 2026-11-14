package bus

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	portal       = "org.freedesktop.portal.Desktop"
	portalPath   = "/org/freedesktop/portal/desktop"
	chooser      = "org.freedesktop.portal.FileChooser"
	request      = "org.freedesktop.portal.Request"
	signalKind   = 4
	choosing     = time.Hour
	matchSerial  = 2
	chooseSerial = 3
)

var ErrNothingChosen = errors.New("nothing was chosen")

type encoder struct {
	data []byte
}

func (e *encoder) align(size int) {
	e.data = padded(e.data, size)
}

func (e *encoder) number(value uint32) {
	e.align(4)
	e.data = binary.LittleEndian.AppendUint32(e.data, value)
}

func (e *encoder) text(value string) {
	e.number(uint32(len(value)))
	e.data = append(append(e.data, value...), 0)
}

func (e *encoder) signature(value string) {
	e.data = append(append(append(e.data, byte(len(value))), value...), 0)
}

func (e *encoder) array(alignment int, fill func()) {
	e.number(0)
	at := len(e.data)
	e.align(alignment)
	start := len(e.data)
	fill()
	binary.LittleEndian.PutUint32(e.data[at-4:at], uint32(len(e.data)-start))
}

func (e *encoder) field(code, kind byte, value string) {
	e.align(8)
	e.data = append(e.data, code, 1, kind, 0)
	if kind == 'g' {
		e.signature(value)
		return
	}
	e.text(value)
}

func methodCallOf(serial uint32, path, face, member, destination, signature string, body []byte) []byte {
	var fields encoder
	fields.field(fieldPath, 'o', path)
	fields.field(fieldInterface, 's', face)
	fields.field(fieldMember, 's', member)
	fields.field(fieldDestination, 's', destination)
	if signature != "" {
		fields.field(fieldSignature, 'g', signature)
	}
	head := binary.LittleEndian.AppendUint32([]byte{'l', methodCall, 0, 1}, uint32(len(body)))
	head = binary.LittleEndian.AppendUint32(head, serial)
	head = binary.LittleEndian.AppendUint32(head, uint32(len(fields.data)))
	return append(padded(append(head, fields.data...), 8), body...)
}

func chooserOptions(token string, kinds []string) []byte {
	var body encoder
	body.text("")
	body.text("Add pictures")
	body.array(8, func() {
		body.align(8)
		body.text("handle_token")
		body.signature("s")
		body.text(token)
		body.align(8)
		body.text("multiple")
		body.signature("b")
		body.number(1)
		body.align(8)
		body.text("filters")
		body.signature("a(sa(us))")
		body.array(8, func() {
			body.align(8)
			body.text("Pictures")
			body.array(8, func() {
				for _, kind := range kinds {
					body.align(8)
					body.number(1)
					body.text(kind)
				}
			})
		})
	})
	return body.data
}

func matchRule() []byte {
	var body encoder
	body.text("type='signal',interface='" + request + "',member='Response'")
	return body.data
}

func ChoosePictures(kinds []string) ([]string, error) {
	path, err := address()
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("unix", path, limit)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(choosing)); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	if err := authenticate(conn, reader); err != nil {
		return nil, err
	}
	token := fmt.Sprintf("solanum%d", os.Getpid())
	outgoing := append([]byte("BEGIN\r\n"), call(1, "Hello")...)
	outgoing = append(outgoing, methodCallOf(matchSerial, daemonPath, daemon, "AddMatch", daemon, "s", matchRule())...)
	outgoing = append(outgoing, methodCallOf(chooseSerial, portalPath, chooser, "OpenFile", portal, "ssa{sv}", chooserOptions(token, kinds))...)
	if _, err := conn.Write(outgoing); err != nil {
		return nil, err
	}
	return chosen(reader)
}

func chosen(reader *bufio.Reader) ([]string, error) {
	for {
		got, err := receive(reader)
		if err != nil {
			return nil, err
		}
		if got.kind == methodError && (got.replyTo == matchSerial || got.replyTo == chooseSerial) {
			return nil, errors.New("the file chooser isn't there")
		}
		if got.kind == signalKind && got.member == "Response" {
			return got.uris()
		}
	}
}

func (m message) uris() ([]string, error) {
	body := &decoder{data: m.body, order: m.order}
	if m.signature != "ua{sv}" {
		return nil, errGarbled
	}
	if body.number() != 0 {
		return nil, ErrNothingChosen
	}
	size := int(body.number())
	body.align(8)
	end := body.at + size
	var found []string
	for !body.bad && body.at < end {
		body.align(8)
		key, kind := body.text(), body.signature()
		if key == "uris" && kind == "as" {
			found = body.strings()
			continue
		}
		body.skip(kind)
	}
	if body.bad {
		return nil, errGarbled
	}
	return paths(found)
}

func paths(uris []string) ([]string, error) {
	var found []string
	for _, uri := range uris {
		parsed, err := url.Parse(uri)
		if err != nil || parsed.Scheme != "file" {
			return nil, fmt.Errorf("the file chooser gave %q, not a file here", uri)
		}
		found = append(found, parsed.Path)
	}
	if len(found) == 0 {
		return nil, ErrNothingChosen
	}
	return found, nil
}

func (d *decoder) strings() []string {
	size := int(d.number())
	end := d.at + size
	var found []string
	for !d.bad && d.at < end {
		found = append(found, d.text())
	}
	return found
}

func complete(signature string) string {
	if signature == "" {
		return ""
	}
	switch signature[0] {
	case 'a':
		return "a" + complete(signature[1:])
	case '(', '{':
		depth := 0
		for index, char := range signature {
			depth += strings.Count("({", string(char)) - strings.Count(")}", string(char))
			if depth == 0 {
				return signature[:index+1]
			}
		}
		return signature
	}
	return signature[:1]
}

func alignment(kind byte) int {
	switch kind {
	case 'y', 'g', 'v':
		return 1
	case 'n', 'q':
		return 2
	case 'x', 't', 'd', '(', '{':
		return 8
	}
	return 4
}

func (d *decoder) skip(signature string) {
	for rest := signature; rest != "" && !d.bad; {
		one := complete(rest)
		d.skipOne(one)
		rest = rest[len(one):]
	}
}

func (d *decoder) skipOne(kind string) {
	switch kind[0] {
	case 'y', 'n', 'q', 'x', 't', 'd':
		size := map[byte]int{'y': 1, 'n': 2, 'q': 2, 'x': 8, 't': 8, 'd': 8}[kind[0]]
		d.align(size)
		d.take(size)
	case 'b', 'i', 'u', 'h':
		d.number()
	case 's', 'o':
		d.text()
	case 'g':
		d.signature()
	case 'v':
		d.skip(d.signature())
	case 'a':
		size := int(d.number())
		d.align(alignment(kind[1]))
		d.take(size)
	case '(', '{':
		d.align(8)
		d.skip(kind[1 : len(kind)-1])
	default:
		d.bad = true
	}
}
