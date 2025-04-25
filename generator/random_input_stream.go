// 原始语言: Java
// 原始文件: RandomInputStream.java
// 目的: 生成随机数据作为上传数据的输入流

package generator

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"math/rand/v2"

	"go.uber.org/zap"

	"github.com/sine-io/sinx/pkg/logger" // 请根据实际模块路径调整
)

// 常量
const BufferSize = 4096 // 4 KB

// RandomInputStream 生成随机数据作为上传数据的输入流
type RandomInputStream struct {
	buffer     []byte
	size       int64
	processed  int64
	hashCheck  bool
	hash       hash.Hash
	hashLen    int
	hashBytes  []byte
	endReached bool // 标记是否已到达流结尾
}

// NewRandomInputStream 创建一个新的随机输入流
func NewRandomInputStream(size int64, random *rand.Rand, isRandom bool, hashCheck bool) *RandomInputStream {
	rs := &RandomInputStream{
		buffer:    make([]byte, BufferSize),
		size:      size,
		hashCheck: hashCheck,
		hash:      md5.New(),
		hashLen:   md5.Size * 2, // 十六进制字符串长度
	}

	// 如果大小太小无法嵌入校验和，禁用完整性检查
	if size <= int64(rs.hashLen) && hashCheck {
		logger.Warn("size is too small to embed checksum, will ignore integrity checking",
			zap.Int64("size", size),
			zap.Int("hash_length", rs.hashLen))
		rs.hashCheck = false
		rs.hash = nil
		rs.hashLen = 0
	}

	// 填充随机内容
	if isRandom {
		for i := range rs.buffer {
			rs.buffer[i] = byte(random.IntN(26) + 'a')
		}
	}

	return rs
}

// Read 实现io.Reader接口，从随机输入流读取数据
func (rs *RandomInputStream) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	// 检查是否已到达流末尾
	if rs.processed >= rs.size {
		return 0, io.EOF
	}

	// 计算本次可读取的最大字节数
	bytesToRead := len(b)
	remaining := rs.size - rs.processed
	if int64(bytesToRead) > remaining {
		bytesToRead = int(remaining)
	}

	// 计算内容部分大小（不包括哈希值）
	contentSize := rs.size
	if rs.hashCheck {
		contentSize -= int64(rs.hashLen)
	}

	// 处理内容部分
	if rs.processed < contentSize {
		// 计算需要读取的内容部分
		contentToRead := bytesToRead
		if rs.processed+int64(contentToRead) > contentSize {
			contentToRead = int(contentSize - rs.processed)
		}

		// 复制数据
		bytesRead := 0
		for bytesRead < contentToRead {
			chunkSize := min(BufferSize, contentToRead-bytesRead)
			copy(b[bytesRead:bytesRead+chunkSize], rs.buffer[:chunkSize])

			// 计算哈希（仅当启用哈希校验时）
			if rs.hashCheck {
				rs.hash.Write(rs.buffer[:chunkSize])
			}

			bytesRead += chunkSize
		}

		rs.processed += int64(bytesRead)

		// 如果刚好读取完内容部分且还有剩余空间，继续读取哈希部分
		if rs.hashCheck && rs.processed == contentSize && bytesRead < bytesToRead {
			hashBytesRead, err := rs.readHashPart(b[bytesRead:], bytesToRead-bytesRead)
			return bytesRead + hashBytesRead, err
		}

		// 检查是否到达流结尾
		if rs.processed >= rs.size {
			return bytesRead, io.EOF
		}

		return bytesRead, nil
	}

	// 处理哈希部分
	if rs.hashCheck && rs.processed >= contentSize {
		return rs.readHashPart(b, bytesToRead)
	}

	return 0, nil
}

// readHashPart 读取哈希部分
func (rs *RandomInputStream) readHashPart(b []byte, bytesToRead int) (int, error) {
	// 计算哈希值（如果尚未计算）
	if rs.hashBytes == nil && !rs.endReached {
		hashSum := rs.hash.Sum(nil)
		rs.hashBytes = []byte(hex.EncodeToString(hashSum))
		rs.endReached = true
	}

	// 计算哈希部分偏移量和可读取的字节数
	contentSize := rs.size - int64(rs.hashLen)
	hashOffset := int(rs.processed - contentSize)
	hashBytesToRead := min(rs.hashLen-hashOffset, bytesToRead)

	// 复制哈希数据
	copy(b[:hashBytesToRead], rs.hashBytes[hashOffset:hashOffset+hashBytesToRead])
	rs.processed += int64(hashBytesToRead)

	// 检查是否到达流结尾
	if rs.processed >= rs.size {
		return hashBytesToRead, io.EOF
	}

	return hashBytesToRead, nil
}

// Len 返回流的总大小
func (rs *RandomInputStream) Len() int64 {
	return rs.size
}

// Buffer 返回内部使用的缓冲区
func (rs *RandomInputStream) Buffer() []byte {
	return rs.buffer
}

// NewReader 创建一个字节读取器，用于读取流内容
func (rs *RandomInputStream) NewReader() *bytes.Reader {
	data := make([]byte, rs.size)

	// 保存当前位置并从头开始读取
	currentPos := rs.processed
	rs.Reset()

	bytesRead, err := io.ReadFull(rs, data)
	if err != nil && err != io.EOF {
		logger.Error("failed to read full data",
			zap.Error(err),
			zap.Int("bytesRead", bytesRead),
			zap.Int64("expectedSize", rs.size))
	}

	// 恢复原来的位置
	rs.processed = currentPos
	return bytes.NewReader(data)
}

// Reset 重置流的状态，使其可以重新读取
func (rs *RandomInputStream) Reset() {
	rs.processed = 0
	rs.endReached = false
	rs.hashBytes = nil
	if rs.hashCheck {
		rs.hash = md5.New()
	}
}
