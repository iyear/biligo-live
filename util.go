package live

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"github.com/andybalholm/brotli"
	"io"
)

func write(l int, n int) []byte {
	b := make([]byte, l)
	switch l {
	case 2:
		binary.BigEndian.PutUint16(b, uint16(n))
	case 4:
		binary.BigEndian.PutUint32(b, uint32(n))
	case 8:
		binary.BigEndian.PutUint64(b, uint64(n))
	}
	return b
}
func encode(ver, op uint8, body []byte) []byte {
	header := make([]byte, 0, wsPackHeaderTotalLen)
	header = append(header, write(wsPackageLen, len(body)+wsPackHeaderTotalLen)...)
	header = append(header, write(wsHeaderLen, wsPackHeaderTotalLen)...)
	header = append(header, write(wsVerLen, int(ver))...)
	header = append(header, write(wsOpLen, int(op))...)
	header = append(header, write(wsSequenceLen, wsHeaderDefaultSequence)...)

	return append(header, body...)
}

// decode 必须确保len(b)>16
func decode(b []byte) (ver uint16, op uint32, body []byte) {
	return binary.BigEndian.Uint16(b[wsPackageLen+wsHeaderLen : wsPackageLen+wsHeaderLen+wsVerLen]),
		binary.BigEndian.Uint32(b[wsPackageLen+wsHeaderLen+wsVerLen : wsPackageLen+wsHeaderLen+wsVerLen+wsOpLen]),
		b[wsPackHeaderTotalLen:]
}
func zlibDe(src []byte) ([]byte, error) {
	var (
		r   io.ReadCloser
		o   bytes.Buffer
		err error
	)
	if r, err = zlib.NewReader(bytes.NewReader(src)); err != nil {
		return nil, err
	}
	if _, err = io.Copy(&o, r); err != nil {
		return nil, err
	}
	return o.Bytes(), nil
}
func brotliDe(src []byte) ([]byte, error) {
	o := new(bytes.Buffer)
	r := brotli.NewReader(bytes.NewReader(src))
	if _, err := io.Copy(o, r); err != nil {
		return nil, err
	}
	return o.Bytes(), nil
}
func brotliEn(src []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	w := brotli.NewWriter(b)

	if _, err := w.Write(src); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
