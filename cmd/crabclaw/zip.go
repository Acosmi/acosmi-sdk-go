package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	acosmi "github.com/acosmi/acosmi-sdk-go"
)

const (
	maxZIPFiles    = 50
	maxZIPTotalMB  = 50
	maxZIPTotalLen = maxZIPTotalMB << 20
)

// extractSkillZIP 安全解压技能 ZIP 到目标目录
func extractSkillZIP(zipData []byte, destDir string) error {
	if destDir == "" {
		return fmt.Errorf("destination directory is empty")
	}

	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	if len(r.File) > maxZIPFiles {
		return fmt.Errorf("zip contains too many files (%d, max %d)", len(r.File), maxZIPFiles)
	}

	// 确保目标目录存在
	cleanDest := filepath.Clean(destDir)
	if err := os.MkdirAll(cleanDest, 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	var actualWritten int64

	for _, f := range r.File {
		// Zip Slip 防护
		cleaned := filepath.Clean(f.Name)
		destPath := filepath.Join(cleanDest, cleaned)

		// 双重校验: 拒绝绝对路径 + 拒绝目录逃逸
		if filepath.IsAbs(f.Name) ||
			!strings.HasPrefix(destPath, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe zip entry: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("create dir %s: %w", cleaned, err)
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("create parent dir: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", cleaned, err)
		}

		// 权限安全: 剥离 setuid/setgid/sticky, 最大 0755
		mode := f.Mode() & 0755

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create file %s: %w", cleaned, err)
		}

		// 限制实际写入量 (不信任声明的 UncompressedSize64, 防止解压炸弹)
		remaining := maxZIPTotalLen - actualWritten
		n, err := io.Copy(outFile, io.LimitReader(rc, remaining+1)) // +1 用于检测超限
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("write file %s: %w", cleaned, err)
		}

		actualWritten += n
		if actualWritten > maxZIPTotalLen {
			return fmt.Errorf("zip total uncompressed size exceeds %dMB", maxZIPTotalMB)
		}
	}

	return nil
}

// packSkillZIP 将生成结果打包为 skill ZIP
func packSkillZIP(result *acosmi.GenerateSkillResult) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// manifest.json
	manifest := map[string]interface{}{
		"name":        result.SkillName,
		"key":         result.SkillKey,
		"description": result.Description,
		"category":    result.Category,
		"timeout":     result.Timeout,
		"version":     "1.0.0",
		"tags":        result.Tags,
	}
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := addZIPFile(w, "manifest.json", manifestData); err != nil {
		return nil, err
	}

	// input-schema.json
	if result.InputSchema != "" {
		if err := addZIPFile(w, "input-schema.json", []byte(result.InputSchema)); err != nil {
			return nil, err
		}
	}

	// output-schema.json
	if result.OutputSchema != "" {
		if err := addZIPFile(w, "output-schema.json", []byte(result.OutputSchema)); err != nil {
			return nil, err
		}
	}

	// README.md
	if result.Readme != "" {
		if err := addZIPFile(w, "README.md", []byte(result.Readme)); err != nil {
			return nil, err
		}
	}

	// skill.md
	if result.SkillMd != "" {
		if err := addZIPFile(w, "skill.md", []byte(result.SkillMd)); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func addZIPFile(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", name, err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write zip entry %s: %w", name, err)
	}
	return nil
}
