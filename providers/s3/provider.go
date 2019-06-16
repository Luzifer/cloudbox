package s3

import (
	"bytes"
	"crypto/md5"
	"hash"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

type Provider struct {
	bucket string
	s3     *s3.S3
}

func New(uri string) (providers.CloudProvider, error) {
	if !strings.HasPrefix(uri, "s3://") {
		return nil, providers.ErrInvalidURI
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid URI specified")
	}

	cfg := aws.NewConfig()
	if u.User != nil {
		user := u.User.Username()
		pass, _ := u.User.Password()
		cfg = cfg.WithCredentials(credentials.NewStaticCredentials(user, pass, ""))
	}

	svc := s3.New(session.Must(session.NewSession(cfg)))

	return &Provider{
		bucket: u.Host,
		s3:     svc,
	}, nil
}

func (p *Provider) Capabilities() providers.Capability {
	return providers.CapBasic | providers.CapAutoChecksum | providers.CapShare
}
func (p *Provider) Name() string                 { return "s3" }
func (p *Provider) GetChecksumMethod() hash.Hash { return md5.New() }

func (p *Provider) DeleteFile(relativeName string) error {
	_, err := p.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(relativeName),
	})

	return errors.Wrap(err, "Unable to delete object")
}

func (p *Provider) GetFile(relativeName string) (providers.File, error) {
	resp, err := p.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(relativeName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch head information")
	}

	return File{
		key:          relativeName,
		lastModified: *resp.LastModified,
		checksum:     strings.Trim(*resp.ETag, `"`),
		size:         uint64(*resp.ContentLength),

		s3Conn: p.s3,
		bucket: p.bucket,
	}, nil
}

func (p *Provider) ListFiles() ([]providers.File, error) {
	var files []providers.File

	err := p.s3.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(p.bucket),
	}, func(out *s3.ListObjectsOutput, lastPage bool) bool {
		for _, obj := range out.Contents {
			files = append(files, File{
				key:          *obj.Key,
				lastModified: *obj.LastModified,
				checksum:     strings.Trim(*obj.ETag, `"`),
				size:         uint64(*obj.Size),

				s3Conn: p.s3,
				bucket: p.bucket,
			})
		}

		return !lastPage
	})

	return files, errors.Wrap(err, "Unable to list objects")
}

func (p *Provider) PutFile(f providers.File) (providers.File, error) {
	body, err := f.Content()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get file reader")
	}
	defer body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, body); err != nil {
		return nil, errors.Wrap(err, "Unable to read source file")
	}

	if _, err = p.s3.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(buf.Bytes()),
		Bucket: aws.String(p.bucket),
		Key:    aws.String(f.Info().RelativeName),
	}); err != nil {
		return nil, errors.Wrap(err, "Unable to write file")
	}

	return p.GetFile(f.Info().RelativeName)
}

func (p *Provider) Share(relativeName string) (string, error) {
	return "", errors.New("Not implemented")
}
