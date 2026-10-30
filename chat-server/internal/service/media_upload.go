package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"

	"github.com/google/uuid"
)

var allowedImageMIMEs = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type mediaUploadSpec struct {
	kind      uint8
	dir       string
	label     string
	sniff     func(header []byte) (mimeType, ext string, ok bool)
	rejectMsg string
}

func (s *MediaService) UploadImage(ctx context.Context, uploader MediaUploader, reader io.Reader, size int64) (*types.MediaResponse, error) {
	return s.upload(ctx, uploader, reader, size, mediaUploadSpec{
		kind:      model.MediaKindImage,
		dir:       "image",
		label:     "图片",
		sniff:     sniffImageMIME,
		rejectMsg: "仅支持 jpg、png、gif、webp 格式图片",
	})
}

func (s *MediaService) UploadVideo(ctx context.Context, uploader MediaUploader, reader io.Reader, size int64) (*types.MediaResponse, error) {
	return s.upload(ctx, uploader, reader, size, mediaUploadSpec{
		kind:      model.MediaKindVideo,
		dir:       "video",
		label:     "视频",
		sniff:     sniffVideoMIME,
		rejectMsg: "仅支持 mp4、mov、webm、3gp 格式视频",
	})
}

func (s *MediaService) UploadAudio(ctx context.Context, uploader MediaUploader, reader io.Reader, size int64) (*types.MediaResponse, error) {
	return s.upload(ctx, uploader, reader, size, mediaUploadSpec{
		kind:      model.MediaKindAudio,
		dir:       "audio",
		label:     "语音",
		sniff:     sniffAudioMIME,
		rejectMsg: "仅支持 m4a、mp3、webm、ogg、wav 格式语音",
	})
}

func (s *MediaService) UploadFile(ctx context.Context, uploader MediaUploader, reader io.Reader, size int64, filename string) (*types.MediaResponse, error) {
	return s.upload(ctx, uploader, reader, size, mediaUploadSpec{
		kind:  model.MediaKindFile,
		dir:   "file",
		label: "文件",
		sniff: func(header []byte) (string, string, bool) {
			return sniffFileMIME(header, filename)
		},
		rejectMsg: "仅支持 pdf、doc、docx、xls、xlsx、ppt、pptx、txt、csv、zip、rar、7z 格式文件",
	})
}

func sniffImageMIME(header []byte) (mimeType, ext string, ok bool) {
	mimeType = http.DetectContentType(header)
	ext, ok = allowedImageMIMEs[mimeType]
	return mimeType, ext, ok
}

func sniffFileMIME(header []byte, filename string) (mimeType, ext string, ok bool) {
	ext = strings.ToLower(path.Ext(strings.TrimSpace(filename)))
	if ext == "" {

		if len(header) >= 4 && string(header[0:4]) == "%PDF" {
			return "application/pdf", ".pdf", true
		}
		if len(header) >= 4 && header[0] == 0x50 && header[1] == 0x4B {
			return "application/zip", ".zip", true
		}
		return "", "", false
	}
	switch ext {
	case ".pdf":
		return "application/pdf", ext, true
	case ".doc":
		return "application/msword", ext, true
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", ext, true
	case ".xls":
		return "application/vnd.ms-excel", ext, true
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ext, true
	case ".ppt":
		return "application/vnd.ms-powerpoint", ext, true
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation", ext, true
	case ".txt":
		return "text/plain", ext, true
	case ".csv":
		return "text/csv", ext, true
	case ".zip":
		return "application/zip", ext, true
	case ".rar":
		return "application/vnd.rar", ext, true
	case ".7z":
		return "application/x-7z-compressed", ext, true
	default:
		return "", "", false
	}
}

func sniffVideoMIME(header []byte) (mimeType, ext string, ok bool) {
	if len(header) >= 12 && string(header[4:8]) == "ftyp" {
		brand := string(header[8:12])
		switch {
		case brand == "qt  ":
			return "video/quicktime", ".mov", true
		case strings.HasPrefix(brand, "3gp"):
			return "video/3gpp", ".3gp", true
		default:
			return "video/mp4", ".mp4", true
		}
	}
	if len(header) >= 4 && header[0] == 0x1A && header[1] == 0x45 && header[2] == 0xDF && header[3] == 0xA3 {
		return "video/webm", ".webm", true
	}
	return "", "", false
}

func sniffAudioMIME(header []byte) (mimeType, ext string, ok bool) {
	if len(header) >= 4 && header[0] == 0x1A && header[1] == 0x45 && header[2] == 0xDF && header[3] == 0xA3 {
		return "audio/webm", ".webm", true
	}
	if len(header) >= 12 && string(header[4:8]) == "ftyp" {
		brand := strings.TrimSpace(string(header[8:12]))
		switch brand {
		case "M4A", "M4B", "M4P", "F4A", "F4B", "mp4a":
			return "audio/mp4", ".m4a", true
		default:
			return "", "", false
		}
	}
	if len(header) >= 4 && string(header[0:4]) == "OggS" {
		return "audio/ogg", ".ogg", true
	}
	if len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WAVE" {
		return "audio/wav", ".wav", true
	}
	if len(header) >= 3 && string(header[0:3]) == "ID3" {
		return "audio/mpeg", ".mp3", true
	}
	if len(header) >= 2 && header[0] == 0xFF && header[1]&0xE0 == 0xE0 {
		return "audio/mpeg", ".mp3", true
	}
	return "", "", false
}

const sniffLen = 512

func (s *MediaService) upload(ctx context.Context, uploader MediaUploader, reader io.Reader, size int64, spec mediaUploadSpec) (*types.MediaResponse, error) {
	limitMB := s.limitMB[spec.kind]
	maxBytes := s.MaxUploadBytes(spec.kind)
	tooLarge := errs.Invalid(fmt.Sprintf("%s大小不能超过 %dMB", spec.label, limitMB))
	if size <= 0 {
		return nil, errs.Invalid("文件不能为空")
	}
	if size > maxBytes {
		return nil, tooLarge
	}

	header := make([]byte, sniffLen)
	n, err := io.ReadFull(reader, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	header = header[:n]
	if len(header) == 0 {
		return nil, errs.Invalid("文件不能为空")
	}

	mimeType, ext, ok := spec.sniff(header)
	if !ok {
		return nil, errs.Invalid(spec.rejectMsg)
	}

	asset := &model.MediaAsset{
		UploaderKind: uploader.Kind,
		UploaderID:   uploader.ID,
		Kind:         spec.kind,
		ObjectKey:    s.objectKey(uploader, spec.dir, ext),
		MimeType:     mimeType,
		Ext:          ext,
		Status:       model.MediaStatusPending,
	}
	if err := s.assetRepo.Create(ctx, asset); err != nil {
		s.log.Error(ctx, "登记媒体失败", logger.Error(err))
		return nil, fmt.Errorf("上传失败，请稍后重试: %w", err)
	}

	counter := &countingReader{r: io.MultiReader(bytes.NewReader(header), reader)}
	body := &limitedReader{r: counter, remaining: maxBytes}
	if err := s.client.Put(ctx, asset.ObjectKey, body, size, mimeType); err != nil {
		s.discard(ctx, asset.ID)
		if errors.Is(err, errReaderTooLarge) {
			return nil, tooLarge
		}
		s.log.Error(ctx, "上传"+spec.label+"失败", logger.Error(err))
		return nil, fmt.Errorf("上传失败，请稍后重试: %w", err)
	}
	marked, err := s.assetRepo.MarkUploaded(ctx, asset.ID, counter.n)
	if err != nil || !marked {
		s.discard(ctx, asset.ID)
		s.log.Error(ctx, "确认媒体上传失败", logger.Uint("mediaId", asset.ID), logger.Error(err))
		return nil, fmt.Errorf("上传失败，请稍后重试: %w", err)
	}

	return &types.MediaResponse{
		ID:        asset.ID,
		URL:       s.client.URL(asset.ObjectKey),
		Kind:      spec.dir,
		MimeType:  mimeType,
		SizeBytes: counter.n,
	}, nil
}

func (s *MediaService) objectKey(uploader MediaUploader, dir, ext string) string {
	owner := fmt.Sprintf("u%d", uploader.ID)
	if uploader.Kind == model.MediaUploaderEmployee {
		owner = fmt.Sprintf("e%d", uploader.ID)
	}
	return path.Join(s.pathPrefix, dir, time.Now().Format("2006/01/02"), owner, uuid.NewString()+ext)
}

func (s *MediaService) discard(ctx context.Context, mediaID uint) {
	ctx = context.WithoutCancel(ctx)
	if err := s.assetRepo.Discard(ctx, mediaID); err != nil {
		s.log.Error(ctx, "放弃上传记录失败", logger.Uint("mediaId", mediaID), logger.Error(err))
	}
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

var errReaderTooLarge = errors.New("文件大小超出限制")

type limitedReader struct {
	r         io.Reader
	remaining int64
}

func (l *limitedReader) Read(p []byte) (int, error) {
	if l.remaining <= 0 {
		var probe [1]byte
		n, err := l.r.Read(probe[:])
		if n > 0 {
			return 0, errReaderTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}
	n, err := l.r.Read(p)
	l.remaining -= int64(n)
	return n, err
}
