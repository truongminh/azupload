package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/caarlos0/env"
	"github.com/google/uuid"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/google/logger"
	_ "github.com/joho/godotenv/autoload" // load .env
)

func main() {
	logger.Init("azupload", true, false, ioutil.Discard)
	s := &server{}
	err := env.Parse(&s.c)
	if err != nil {
		logger.Fatalf("load env error %s", err.Error())
	}
	s.c.Azure.Prefix = strings.TrimSuffix(strings.TrimPrefix(s.c.Azure.Prefix, "/"), "/")
	s.c.BaseURL = strings.TrimSuffix(s.c.BaseURL, "/")
	az := s.c.Azure
	s.az = newAZContainer(az.Name, az.Key, az.Container)
	addr := fmt.Sprintf(":%d", s.c.HTTPPORT)
	logger.Infof("Listening on %s", addr)
	http.ListenAndServe(addr, s)
}

type Config struct {
	HTTPPORT int `env:"HTTP_PORT" envDefault:"3000"`
	Azure    struct {
		Name      string `env:"AZURE_BLOB_ACCOUNT_NAME,required"`
		Key       string `env:"AZURE_BLOB_ACCOUNT_KEY,required"`
		Container string `env:"AZURE_BLOB_CONTAINER,required"`
		Prefix    string `env:"AZURE_BLOB_PREFIX,required"`
	}
	BaseURL string `env:"BASE_URL"`
}

func newAZContainer(accountName string, accountKey string, containerName string) azblob.ContainerURL {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		logger.Fatal("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName),
	)
	return azblob.NewContainerURL(*URL, p)
}

type server struct {
	c  Config
	az azblob.ContainerURL
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.get(w, r)
	} else if r.Method == "POST" {
		s.post(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	blobName := r.URL.Path[1:]
	blobURL := s.az.NewBlockBlobURL(blobName)
	ctx := r.Context()
	get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		if storageErr, ok := err.(azblob.StorageError); ok {
			switch storageErr.ServiceCode() {
			case azblob.ServiceCodeResourceNotFound, azblob.ServiceCodeBlobNotFound:
				http.NotFound(w, r)
				return
			case azblob.ServiceCodeInvalidURI:
				http.Error(w, "invalid file uri", http.StatusBadRequest)
				return
			default:
				logger.Errorf("download blob error %s", err.Error())
			}
		}
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	reader := get.Body(azblob.RetryReaderOptions{MaxRetryRequests: 3})
	io.Copy(w, reader)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func (s *server) post(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "bad form file", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	blobName := s.c.Azure.Prefix + singleJoiningSlash(r.URL.Path, uuid.New().String()) + "/" + header.Filename
	logger.Infof("uploading to blob %s", blobName)
	blobURL := s.az.NewBlockBlobURL(blobName)
	_, err = azblob.UploadStreamToBlockBlob(
		ctx, file, blobURL, azblob.UploadStreamToBlockBlobOptions{
			MaxBuffers: 100,     // 100 MB
			BufferSize: 1 << 20, // 1MB
		},
	)
	if err != nil {
		logger.Errorf("storage error %s", err.Error())
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	url := s.c.BaseURL + "/" + blobName
	res := map[string]interface{}{
		"url": url,
	}
	json.NewEncoder(w).Encode(res)
}
