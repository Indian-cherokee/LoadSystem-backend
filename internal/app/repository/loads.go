package repository

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"web/internal/app/ds"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) GetAllLoads() ([]ds.Loads, error) {
	var loads []ds.Loads

	err := r.db.Find(&loads).Error
	if err != nil {
		return nil, err
	}

	return loads, nil
}

func (r *Repository) SearchLoadsByName(title string) ([]ds.Loads, error) {
	var loads []ds.Loads
	err := r.db.Where("load_title ILIKE ?", "%"+title+"%").Find(&loads).Error
	if err != nil {
		return nil, err
	}
	return loads, nil
}

func (r *Repository) GetLoadByID(id uint) (*ds.Loads, error) {
	var load ds.Loads
	err := r.db.First(&load, id).Error
	if err != nil {
		return nil, err
	}
	return &load, nil
}

func (r *Repository) GetLoadsFiltered(search string) ([]ds.Loads, error) {
	var loads []ds.Loads
	query := r.db

	if search != "" {
		query = query.Where("load_title ILIKE ?", "%"+search+"%")
	}

	err := query.Find(&loads).Error
	if err != nil {
		return nil, err
	}
	return loads, nil
}

func (r *Repository) CreateLoad(load *ds.Loads) error {
	return r.db.Create(load).Error
}

func (r *Repository) UpdateLoad(id uint, req ds.LoadUpdateRequest) (*ds.Loads, error) {
	var load ds.Loads
	if err := r.db.First(&load, id).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.LoadTitle != nil {
		updates["load_title"] = *req.LoadTitle
	}
	if req.LoadDescription != nil {
		updates["load_description"] = *req.LoadDescription
	}
	if req.Normative != nil {
		updates["normative"] = *req.Normative
	}
	if req.LoadCategory != nil {
		updates["load_category"] = *req.LoadCategory
	}
	if req.ReliabilityCoefficient != nil {
		updates["reliability_coefficient"] = *req.ReliabilityCoefficient
	}

	if err := r.db.Model(&load).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &load, nil
}

func (r *Repository) DeleteLoad(id uint) error {
	var load ds.Loads
	var imageURLToDelete string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&load, id).Error; err != nil {
			return err
		}
		if load.LoadImage != nil {
			imageURLToDelete = *load.LoadImage
		}
		if err := tx.Delete(&ds.Loads{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	if imageURLToDelete != "" {
		parsedURL, err := url.Parse(imageURLToDelete)
		if err != nil {
			log.Printf("ERROR: could not parse image URL for deletion: %v", err)
			return nil
		}

		objectName := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", r.bucketName))

		err = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("ERROR: failed to delete object '%s' from MinIO: %v", objectName, err)
		}
	}

	return nil
}

func (r *Repository) UploadLoadImage(id uint, file *multipart.FileHeader) (string, error) {
	var finalImageURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var load ds.Loads
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&load, id).Error; err != nil {
			return fmt.Errorf("load with id %d not found: %w", id, err)
		}

		const imagePathPrefix = "Images/"

		if load.LoadImage != nil && *load.LoadImage != "" {
			oldImageURL, err := url.Parse(*load.LoadImage)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		fileName := filepath.Base(file.Filename)
		objectName := imagePathPrefix + fileName

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		_, err = r.minioClient.PutObject(context.Background(), r.bucketName, objectName, fileReader, file.Size, minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		})

		if err != nil {
			return fmt.Errorf("failed to upload to minio: %w", err)
		}

		imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.bucketName, objectName)

		if err := tx.Model(&load).Update("load_image", imageURL).Error; err != nil {
			return fmt.Errorf("failed to update load image url in db: %w", err)
		}

		finalImageURL = imageURL
		return nil
	})
	if err != nil {
		return "", err
	}
	return finalImageURL, nil
}

func (r *Repository) AddLoadToDraft(userID, loadID uint) error {
	draft, err := r.GetDraftLoadSession(userID)
	if err != nil {

		newDraft := ds.LoadSession{
			CreatorID:    userID,
			Status:       ds.StatusDraft,
			CreationDate: time.Now(),
		}
		if err := r.CreateLoadSession(&newDraft); err != nil {
			return err
		}
		draft = &newDraft
	}

	return r.AddLoadToLoadSession(draft.ID, loadID)
}
