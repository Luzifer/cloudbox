package s3

import (
	"bytes"
	"fmt"
	"hash"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

type File struct {
	key          string
	lastModified time.Time
	checksum     string
	size         uint64

	s3Conn *s3.S3
	bucket string
	prefix string
}

func (f File) Info() providers.FileInfo {
	return providers.FileInfo{
		RelativeName: strings.Trim(strings.TrimPrefix(f.key, f.prefix), "/"),
		LastModified: f.lastModified,
		Checksum:     f.checksum,
		Size:         f.size,
	}
}

func (f File) Checksum(h hash.Hash) (string, error) {
	cont, err := f.Content()
	if err != nil {
		return "", errors.Wrap(err, "Unable to get file content")
	}
	defer cont.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, cont); err != nil {
		return "", errors.Wrap(err, "Unable to read file content")
	}

	return fmt.Sprintf("%x", h.Sum(buf.Bytes())), nil
}

func (f File) Content() (io.ReadCloser, error) {
	resp, err := f.s3Conn.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.key),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get file")
	}

	return resp.Body, nil
}
