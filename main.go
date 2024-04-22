package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

func main() {
	// create new fiber instance  and use across whole app
	app := fiber.New()

	// middleware to allow all clients to communicate using http and allow cors
	app.Use(cors.New())

	// serve  images from images directory prefixed with /images
	// i.e http://localhost:4000/images/someimage.webp

	// app.Static("/images", "./images")

	app.Get("/images", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"images": uploader.ListFiles(),
		})
	})

	// handle image uploading using post request

	app.Post("/", handleFileupload)

	// delete uploaded image by providing unique image name

	app.Delete("/:imageName", handleDeleteImage)

	// start dev server on port 4000

	log.Fatal(app.Listen(":4000"))
}

func handleFileupload(c *fiber.Ctx) error {
	// parse incomming image file

	file, err := c.FormFile("image")
	if err != nil {
		log.Println("image upload error --> ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})

	}

	// generate new uuid for image name
	uniqueId := uuid.New()

	// remove "- from imageName"

	filename := strings.Replace(uniqueId.String(), "-", "", -1)

	// extract image extension from original file filename

	fileExt := strings.Split(file.Filename, ".")[1]

	// generate image from filename and extension
	image := fmt.Sprintf("%s.%s", filename, fileExt)

	// save image to ./images dir
	filepath := fmt.Sprintf("./images/%s", image)
	err = c.SaveFile(file, fmt.Sprintf("./images/%s", image))
	if err != nil {
		log.Println("image save error --> ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// generate image url to serve to client using CDN

	imageUrl := fmt.Sprintf("http://localhost:4000/images/%s", image)

	fileObj, err := os.Open(filepath)
	if err != nil {
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	CreateImageInformation(filename)

	err = uploader.UploadFile(fileObj, filename)
	if err != nil {
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// create meta data and send to client

	data := map[string]interface{}{
		"imageName": image,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": data})
}

func handleDeleteImage(c *fiber.Ctx) error {
	// extract image name from params
	imageName := c.Params("imageName")

	// delete image from ./images
	err := os.Remove(fmt.Sprintf("./images/%s", imageName))
	if err != nil {
		log.Println(err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server Error", "data": nil})
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image deleted successfully", "data": nil})
}

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

var uploader *ClientUploader

const (
	SCOPE_READ_ONLY  = "https://www.googleapis.com/auth/devstorage.read_only"
	SCOPE_READ_WRITE = "https://www.googleapis.com/auth/devstorage.read_write"
)

func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", secretFile) // FILL IN WITH YOUR FILE PATH
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	uploader = &ClientUploader{
		cl:         client,
		bucketName: bucketName,
		projectID:  projectID,
		uploadPath: "test-files/",
	}
}

type imageInformation struct {
	Link string `json:"link"`
	// Name    string    `json:"name"`
	Created time.Time `json:"created"`
}

func (c *ClientUploader) ListFiles() []imageInformation {
	fileInfos := []imageInformation{}
	// files := []string{}
	it := c.cl.Bucket(c.bucketName).Objects(context.TODO(), nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// return fmt.Errorf("Bucket(%q).Objects: %w", bucket, err)
		}

		imageInfo, err := queries.GetInformation(context.Background(), attrs.Name)
		if err != nil {
			log.Println(err)
		}

		fileInfos = append(fileInfos, imageInformation{
			Link:    attrs.Name,
			Created: imageInfo.Created,
		})
	}

	return fileInfos
}

func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.cl.Bucket(c.bucketName).Object(object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func listFiles(w io.Writer, bucket string) error {
	// bucket := "bucket-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	it := client.Bucket(bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Bucket(%q).Objects: %w", bucket, err)
		}
		fmt.Fprintln(w, attrs.Name)
	}
	return nil
}
