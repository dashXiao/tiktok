package service

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/ozline/tiktok/config"
	"github.com/ozline/tiktok/kitex_gen/video"
)

func (s *VideoService) UploadCover(req *video.UploadVideoRequest, coverName string) (err error) {
	// 用户上传了封面，直接存储
	if len(req.CoverFile) > 0 {
		return s.bucket.PutObject(config.OSS.MainDirectory+"/"+coverName, bytes.NewReader(req.CoverFile))
	}

	var imageBuffer bytes.Buffer // 接收图像输出

	// 使用 stdin/stdout 管道截取首帧，避免命名管道的阻塞问题
	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-vframes", "1", "-f", "image2pipe", "-vcodec", "png", "pipe:1")
	cmd.Stdin = bytes.NewReader(req.VideoFile)
	cmd.Stdout = &imageBuffer

	if err = cmd.Run(); err != nil {
		return err
	}

	if imageBuffer.Len() == 0 {
		return errors.New("cover generation failed: empty output")
	}

	return s.bucket.PutObject(config.OSS.MainDirectory+"/"+coverName, bytes.NewReader(imageBuffer.Bytes()))
}
