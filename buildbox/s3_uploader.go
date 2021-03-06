package buildbox

import (
  "io/ioutil"
  "github.com/crowdmob/goamz/s3"
  "github.com/crowdmob/goamz/aws"
  "os"
  "fmt"
  "strings"
  "errors"
)

type S3Uploader struct {
  // The destination which includes the S3 bucket name
  // and the path.
  // s3://my-bucket-name/foo/bar
  Destination string

  // The S3 Bucket we're uploading these files to
  Bucket *s3.Bucket
}

func (u *S3Uploader) Setup(destination string) (error) {
  u.Destination = destination

  // Setup the AWS authentication
  auth, err := aws.EnvAuth()
  if err != nil {
    fmt.Printf("Error loading AWS credentials: %s", err)
    os.Exit(1)
  }

  // Decide what region to use
  // https://github.com/crowdmob/goamz/blob/master/aws/regions.go
  // I think S3 defaults to us-east-1
  regionName := "us-east-1"
  if os.Getenv("AWS_DEFAULT_REGION") != "" {
    regionName = os.Getenv("AWS_DEFAULT_REGION")
  }

  // Check to make sure the region exists
  region, ok := aws.Regions[regionName]
  if ok == false {
    return errors.New("Unknown AWS Region `" + regionName + "`")
  }

  // Find the bucket
  s3 := s3.New(auth, region)
  bucket := s3.Bucket(u.bucketName())

  // If the list doesn't return an error, then we've got our
  // bucket
  _, err = bucket.List("", "", "", 0)
  if err != nil {
    return errors.New("Could not find bucket `" + u.bucketName() + " in region `" + region.Name + "` (" + err.Error() + ")")
  }

  u.Bucket = bucket

  return nil
}

func (u *S3Uploader) URL(artifact *Artifact) (string) {
  return "http://" + u.bucketName() + ".s3.amazonaws.com/" + u.artifactPath(artifact)
}

func (u *S3Uploader) Upload(artifact *Artifact) (error) {
  Perms := s3.ACL("public-read")

  data, err := ioutil.ReadFile(artifact.AbsolutePath)
  if err != nil {
    return errors.New("Failed to read file " + artifact.AbsolutePath + " (" + err.Error() + ")")
  }

  err = u.Bucket.Put(u.artifactPath(artifact), data, artifact.MimeType(), Perms, s3.Options{})
  if err != nil {
    return errors.New("Failed to PUT file " + u.artifactPath(artifact) + " (" + err.Error() + ")")
  }

  return nil
}

// func (u S3Uploader) Download(file string, bucket *s3.Bucket, path string) {
//   data, err := bucket.Get(path)
//   if err != nil {
//     panic(err.Error())
//   }
//   perms := os.FileMode(0644)
//
//   err = ioutil.WriteFile(file, data, perms)
//   if err != nil {
//     panic(err.Error())
//   }
// }

func (u *S3Uploader) artifactPath(artifact *Artifact) (string) {
  parts := []string{u.bucketPath(), artifact.Path}

  return strings.Join(parts, "/")
}

func (u *S3Uploader) bucketPath() string {
  return strings.Join(u.destinationParts()[1:len(u.destinationParts())], "/")
}

func (u *S3Uploader) bucketName() (string) {
  return u.destinationParts()[0]
}

func (u *S3Uploader) destinationParts() []string {
  trimmed_string := strings.TrimLeft(u.Destination, "s3://")

  return strings.Split(trimmed_string, "/")
}
