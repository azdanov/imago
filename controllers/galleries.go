package controllers

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/azdanov/imago/context"
	"github.com/azdanov/imago/models"
	"github.com/go-chi/chi/v5"
)

const maxFileSize = 5 << 20 // 5 MB

type Galleries struct {
	Templates struct {
		New  Template
		Edit Template
		Show Template
		List Template
	}
	GalleryService *models.GalleryService
}

func NewGalleries(gs *models.GalleryService) *Galleries {
	return &Galleries{
		GalleryService: gs,
	}
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")

	g.Templates.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		UserID int
		Title  string
	}
	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")

	if data.Title == "" {
		vals := url.Values{
			models.NotificationError: {"Title is required"},
		}
		http.Redirect(w, r, "/galleries/new?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Failed to create gallery"},
		}
		http.Redirect(w, r, "/galleries/new?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	galleryID := strconv.Itoa(gallery.ID)

	http.Redirect(w, r, "/galleries/"+galleryID+"/edit", http.StatusSeeOther)
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	data, err := g.fetchGalleryData(r)
	if err != nil {
		http.Redirect(w, r, "/galleries?"+err.Encode(), http.StatusSeeOther)
		return
	}

	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Invalid gallery ID"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")

	if data.Title == "" {
		vals := url.Values{
			models.NotificationError: {"Title is required"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	gallery, err := g.GalleryService.ByID(galleryID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Gallery not found"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	if gallery.UserID != context.User(r.Context()).ID {
		vals := url.Values{
			models.NotificationError: {"You do not have permission to edit this gallery"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	gallery.Title = data.Title

	err = g.GalleryService.Update(gallery)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Failed to update gallery"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	vals := url.Values{
		models.NotificationSuccess: {"Gallery updated successfully"},
	}
	http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	data, err := g.fetchGalleryData(r)
	if err != nil {
		http.Redirect(w, r, "/galleries?"+err.Encode(), http.StatusSeeOther)
		return
	}

	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) fetchGalleryData(r *http.Request) (struct {
	ID     int
	Title  string
	Images []struct {
		Filename        string
		EscapedFilename string
	}
}, url.Values,
) {
	var data struct {
		ID     int
		Title  string
		Images []struct {
			Filename        string
			EscapedFilename string
		}
	}

	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return data, url.Values{
			models.NotificationError: {"Invalid gallery ID"},
		}
	}

	gallery, err := g.GalleryService.ByID(galleryID)
	if err != nil {
		return data, url.Values{
			models.NotificationError: {"Gallery not found"},
		}
	}

	data.ID = gallery.ID
	data.Title = gallery.Title

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		return data, url.Values{
			models.NotificationError: {"Failed to retrieve images"},
		}
	}

	for _, image := range images {
		data.Images = append(data.Images, struct {
			Filename        string
			EscapedFilename string
		}{
			Filename:        image.Filename,
			EscapedFilename: url.PathEscape(image.Filename),
		})
	}

	return data, nil
}

func (g Galleries) List(w http.ResponseWriter, r *http.Request) {
	galleries, err := g.GalleryService.ByUserID(context.User(r.Context()).ID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Failed to retrieve galleries"},
		}
		http.Redirect(w, r, "/?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	var data struct {
		Galleries []models.Gallery
	}
	data.Galleries = galleries

	g.Templates.List.Execute(w, r, data)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Invalid gallery ID"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	gallery, err := g.GalleryService.ByID(galleryID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Gallery not found"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	if gallery.UserID != context.User(r.Context()).ID {
		vals := url.Values{
			models.NotificationError: {"You do not have permission to delete this gallery"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	err = g.GalleryService.Delete(galleryID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Failed to delete gallery"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusSeeOther)
}

func (g Galleries) UploadImage(w http.ResponseWriter, r *http.Request) {
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Invalid gallery ID"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	gallery, err := g.GalleryService.ByID(galleryID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Gallery not found"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	err = r.ParseMultipartForm(maxFileSize)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Failed to parse form. Max file size is " +
				strconv.Itoa(maxFileSize) + " bytes."},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	fileHeaders := r.MultipartForm.File["images"]
	for _, fileHeader := range fileHeaders {
		var file multipart.File
		file, err = fileHeader.Open()
		if err != nil {
			vals := url.Values{
				models.NotificationError: {"Failed to open file"},
			}
			http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
			return
		}

		defer file.Close()

		err = g.GalleryService.CreateImage(gallery.ID, fileHeader.Filename, file)
		if err != nil {
			var fileErr models.FileError
			if errors.As(err, &fileErr) {
				vals := url.Values{
					models.NotificationError: {fmt.Sprintf(`%v has an invalid content type or extension. 
					Only %s files can be uploaded.`, fileHeader.Filename, g.GalleryService.Extensions())},
				}
				http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
				return
			}

			vals := url.Values{
				models.NotificationError: {"Failed to upload image"},
			}
			http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
			return
		}
	}

	vals := url.Values{
		models.NotificationSuccess: {"Image uploaded successfully"},
	}
	http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")

	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Invalid gallery ID"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	image, err := g.GalleryService.Image(galleryID, filename)
	if err != nil {
		log.Printf("retrieving image: %v", err)
		if errors.Is(err, models.ErrNotFound) {
			vals := url.Values{
				models.NotificationError: {"Image not found"},
			}
			http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
			return
		}
		vals := url.Values{
			models.NotificationError: {"Something went wrong"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, image.Path)
}

func (g Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Invalid gallery ID"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	gallery, err := g.GalleryService.ByID(galleryID)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Gallery not found"},
		}
		http.Redirect(w, r, "/galleries?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	filename := chi.URLParam(r, "filename")
	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		vals := url.Values{
			models.NotificationError: {"Failed to delete image"},
		}
		http.Redirect(w, r, "/galleries/"+strconv.Itoa(gallery.ID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	vals := url.Values{
		models.NotificationSuccess: {"Image deleted successfully"},
	}
	http.Redirect(w, r, "/galleries/"+strconv.Itoa(galleryID)+"/edit?"+vals.Encode(), http.StatusSeeOther)
}
