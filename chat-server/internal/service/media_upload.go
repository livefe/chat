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

func (s *MediaService) UploadImage(ctx context.Context, userID uint, reader io.Reader, size int64) (*types.MediaResponse, error) {
	return s.upload(ctx, userID, model.MediaKindImage, "image", reader, size,
		int64(s.maxSizeMB)*1024*1024, s.maxSizeMB, "图片",
		func(header []byte) (string, string, bool) {
			mimeType := http.DetectContentType(header)
			ext, ok := allowedImageMIMEs[mimeType]
			return mimeType, ext, ok
		},
		"仅支持 jpg、png、gif、webp 格式图片",
		"上传图片失败",
	)
}

func (s *MediaService) UploadVideo(ctx context.Context, userID uint, reader io.Reader, size int64) (*types.MediaResponse, error) {
	return s.upload(ctx, userID, model.MediaKindVideo, "video", reader, size,
		int64(s.maxVideoSizeMB)*1024*1024, s.maxVideoSizeMB, "视频",
		sniffVideoMIME,
		"仅支持 mp4、mov、webm、3gp 格式视频",
		"上传视频失败",
	)
}

func (s *MediaService) UploadAudio(ctx context.Context, userID uint, reader io.Reader, size int64) (*types.MediaResponse, error) {
	return s.upload(ctx, userID, model.MediaKindAudio, "audio", reader, size,
		int64(s.maxAudioSizeMB)*1024*1024, s.maxAudioSizeMB, "语音",
		sniffAudioMIME,
		"仅支持 m4a、mp3、webm、ogg、wav 格式语音",
		"上传语音失败",
	)
}

func (s *MediaService) UploadFile(ctx context.Context, userID uint, reader io.Reader, size int64, filename string) (*types.MediaResponse, error) {
	const maxFileMB = 20
	return s.upload(ctx, userID, model.MediaKindFile, "file", reader, size,
		int64(maxFileMB)*1024*1024, maxFileMB, "文件",
		func(header []byte) (string, string, bool) {
			return sniffFileMIME(header, filename)
		},
		"仅支持 pdf、doc、docx、xls、xlsx、ppt、pptx、txt、csv、zip、rar、7z 格式文件",
		"上传文件失败",
	)
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

func (s *MediaService) upload(
	ctx context.Context,
	userID uint,
	kind uint8,
	kindDir string,
	reader io.Reader,
	size int64,
	maxBytes int64,
	maxMB int,
	label string,
	sniff func([]byte) (string, string, bool),
	rejectMsg string,
	logMsg string,
) (*types.MediaResponse, error) {
	if size > maxBytes {
		return nil, errs.Invalid(fmt.Sprintf("%s大小不能超过 %dMB", label, maxMB))
	}
	if size <= 0 {
		return nil, errs.Invalid("文件不能为空")
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

	mimeType, ext, ok := sniff(header)
	if !ok {
		return nil, errs.Invalid(rejectMsg)
	}

	counter := &countingReader{r: io.MultiReader(bytes.NewReader(header), reader)}
	body := &limitedReader{r: counter, remaining: maxBytes}
	objectKey := path.Join(s.pathPrefix, kindDir, time.Now().Format("2006/01/02"), fmt.Sprintf("%d", userID), uuid.NewString()+ext)

	url, err := s.client.Upload(objectKey, body, mimeType)
	if err != nil {
		if errors.Is(err, errReaderTooLarge) {
			return nil, errs.Invalid(fmt.Sprintf("%s大小不能超过 %dMB", label, maxMB))
		}
		s.log.Error(ctx, logMsg, logger.Error(err))
		return nil, fmt.Errorf("上传失败，请稍后重试: %w", err)
	}

	asset := &model.MediaAsset{
		UserID:    userID,
		Kind:      kind,
		ObjectKey: objectKey,
		URL:       url,
		MimeType:  mimeType,
		Ext:       ext,
		SizeBytes: counter.n,
		Status:    model.MediaStatusUploaded,
	}
	if err := s.assetRepo.Create(ctx, asset); err != nil {
		_ = s.client.Delete(ctx, objectKey)
		s.log.Error(ctx, "保存媒体记录失败", logger.Error(err))
		return nil, fmt.Errorf("上传失败，请稍后重试: %w", err)
	}

	return &types.MediaResponse{
		ID:        asset.ID,
		URL:       asset.URL,
		Key:       asset.ObjectKey,
		Kind:      kindDir,
		MimeType:  asset.MimeType,
		SizeBytes: asset.SizeBytes,
	}, nil
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
		return 0, errReaderTooLarge
	}
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}
	n, err := l.r.Read(p)
	l.remaining -= int64(n)
	return n, err
}
