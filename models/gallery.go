package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Gallery struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type Image struct {
	GalleryID int
	Path      string
	Filename  string
}

type GalleryService struct {
	DB       *sql.DB
	ImageDir string
}

func NewGalleryService(db *sql.DB) *GalleryService {
	return &GalleryService{
		DB: db,
	}
}

func (s *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		Title:     title,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	row := s.DB.QueryRow(`INSERT INTO galleries (title, user_id, created_at) VALUES ($1, $2, $3) RETURNING id;`,
		gallery.Title, gallery.UserID, gallery.CreatedAt)

	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}

	return &gallery, nil
}

func (s *GalleryService) ByID(id int) (*Gallery, error) {
	gallery := Gallery{}

	query := `SELECT id, user_id, title, created_at FROM galleries WHERE id = $1;`
	err := s.DB.QueryRow(query, id).Scan(&gallery.ID, &gallery.UserID, &gallery.Title, &gallery.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallery by id: %w", err)
	}

	return &gallery, nil
}

func (s *GalleryService) ByUserID(userID int) ([]Gallery, error) {
	query := `SELECT id, user_id, title, created_at FROM galleries WHERE user_id = $1;`
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query galleries by user id: %w", err)
	}
	defer rows.Close()

	var galleries []Gallery

	for rows.Next() {
		var gallery Gallery
		err = rows.Scan(&gallery.ID, &gallery.UserID, &gallery.Title, &gallery.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan gallery row: %w", err)
		}
		galleries = append(galleries, gallery)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gallery rows: %w", err)
	}

	return galleries, nil
}

func (s *GalleryService) Update(gallery *Gallery) error {
	query := `UPDATE galleries SET title = $1 WHERE id = $2;`
	_, err := s.DB.Exec(query, gallery.Title, gallery.ID)
	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}

	return nil
}

func (s *GalleryService) Delete(id int) error {
	galleryDir := s.galleryDir(id)
	err := os.RemoveAll(galleryDir)
	if err != nil {
		return fmt.Errorf("delete gallery images: %w", err)
	}

	query := `DELETE FROM galleries WHERE id = $1;`
	_, err = s.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}

	return nil
}

func (s *GalleryService) Extensions() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

func (s *GalleryService) galleryDir(id int) string {
	imagesDir := s.ImageDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery_%d", id))
}

func hasExtension(file string, extensions []string) bool {
	for _, ext := range extensions {
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}

func (s *GalleryService) imageContentTypes() []string {
	return []string{"image/png", "image/jpg", "image/gif"}
}

func (s *GalleryService) Images(galleryID int) ([]Image, error) {
	globPattern := filepath.Join(s.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("retrieving gallery images: %w", err)
	}

	var images []Image
	for _, file := range allFiles {
		if hasExtension(file, s.Extensions()) {
			images = append(images, Image{
				GalleryID: galleryID,
				Path:      file,
				Filename:  filepath.Base(file),
			})
		}
	}
	return images, nil
}

func (s *GalleryService) Image(galleryID int, filename string) (Image, error) {
	imagePath := filepath.Clean(filepath.Join(s.galleryDir(galleryID), filename))

	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("querying for image: %w", err)
	}

	image := Image{
		Filename:  filename,
		GalleryID: galleryID,
		Path:      imagePath,
	}

	return image, nil
}

func (s *GalleryService) DeleteImage(galleryID int, filename string) error {
	image, err := s.Image(galleryID, filename)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	err = os.Remove(image.Path)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}

func (s *GalleryService) CreateImage(galleryID int, filename string, contents io.ReadSeeker) error {
	err := checkContentType(contents, s.imageContentTypes())
	if err != nil {
		return fmt.Errorf("creating image %v: %w", filename, err)
	}

	err = checkExtension(filename, s.Extensions())
	if err != nil {
		return fmt.Errorf("creating image %v: %w", filename, err)
	}

	galleryDir := s.galleryDir(galleryID)

	err = os.MkdirAll(galleryDir, 0o750)
	if err != nil {
		return fmt.Errorf("creating gallery-%d images directory: %w", galleryID, err)
	}

	imagePath := filepath.Clean(filepath.Join(galleryDir, filename))

	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("creating image file: %w", err)
	}

	defer dst.Close()

	_, err = io.Copy(dst, contents)
	if err != nil {
		return fmt.Errorf("copying contents to image: %w", err)
	}

	return nil
}
