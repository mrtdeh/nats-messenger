package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	Bucket = "files"
)

func uploadFile(js nats.JetStreamContext, objectName, filePath string) error {
	// Note: Before creating an object store, we need to make sure that the object store does not already exist
	err := js.DeleteObjectStore(Bucket)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return fmt.Errorf("err000 : %v", err)
	}
	// Creating a new object store
	objStore, err := js.CreateObjectStore(&nats.ObjectStoreConfig{
		Bucket: Bucket,
		TTL:    time.Minute * 2,
	})
	if err != nil {
		return fmt.Errorf("err1 : %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("err2 : %v", err)
	}
	defer f.Close()

	info, err := objStore.Put(&nats.ObjectMeta{Name: objectName}, f)
	if err != nil {
		return fmt.Errorf("err3 : %v", err)
	}
	fmt.Println("Uploaded:", info)
	return nil
}

func downloadFile(js nats.JetStreamContext, objectName, destPath string) error {
	objStore, err := js.ObjectStore(Bucket)
	if err != nil {
		return fmt.Errorf("err1 : %v", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("err4 : %v", err)
	}
	defer out.Close()

	obj, err := objStore.Get(objectName)
	if err != nil {
		return fmt.Errorf("err5 : %v", err)
	}

	_, err = io.Copy(out, obj)
	if err != nil {
		return fmt.Errorf("err6 : %v", err)
	}
	fmt.Println("Downloaded OK")
	return nil
}
