// examples/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zhangshuanlai/gtm-storage/client"
	// 替换为实际的模块路径
)

func main() {
	// 创建客户端
	client := client.NewClient(client.ClientOptions{
		BaseURL: "http://127.0.0.1:3000", // 替换为实际的服务器地址
		APIKey:  "sssss",                 // 替换为实际的API密钥
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// 示例1: 创建存储桶
	fmt.Println("=== 创建存储桶 ===")
	bucketName := "test-bucket"
	if err := client.MakeBucket(ctx, bucketName); err != nil {
		log.Printf("创建存储桶失败: %v", err)
	} else {
		fmt.Printf("存储桶 %s 创建成功\n", bucketName)
	}

	// 示例2: 上传文件
	fmt.Println("\n=== 上传文件 ===")
	fileContent := strings.NewReader("Hello, GTM Storage!")
	result, err := client.PutObject(ctx, bucketName, "test-file", fileContent, "test.txt")
	if err != nil {
		log.Printf("上传文件失败: %v", err)
	} else {
		fmt.Printf("文件上传成功:\n")
		fmt.Printf("  Key: %s\n", result.Key)
		fmt.Printf("  ETag: %s\n", result.ETag)
		fmt.Printf("  预览URL: %s\n", result.PreviewURL)
		if result.ThumbnailURL != "" {
			fmt.Printf("  缩略图URL: %s\n", result.ThumbnailURL)
		}
	}

	// 示例3: 上传图片文件
	fmt.Println("\n=== 上传图片文件 ===")
	imageFile, err := os.Open("example.jpg") // 确保文件存在
	if err == nil {
		defer imageFile.Close()
		result, err := client.PutObject(ctx, bucketName, "my-image", imageFile, "example.jpg")
		if err != nil {
			log.Printf("上传图片失败: %v", err)
		} else {
			fmt.Printf("图片上传成功:\n")
			fmt.Printf("  预览URL: %s\n", result.PreviewURL)
			fmt.Printf("  缩略图URL: %s\n", result.ThumbnailURL)
		}
	}

	// 示例4: 获取对象元数据
	fmt.Println("\n=== 获取对象元数据 ===")
	objInfo, err := client.HeadObject(ctx, bucketName, "test-file")
	if err != nil {
		log.Printf("获取对象元数据失败: %v", err)
	} else {
		fmt.Printf("对象元数据:\n")
		fmt.Printf("  Key: %s\n", objInfo.Key)
		fmt.Printf("  ContentType: %s\n", objInfo.ContentType)
		fmt.Printf("  Size: %d bytes\n", objInfo.Size)
		fmt.Printf("  LastModified: %s\n", objInfo.LastModified.Format(time.RFC3339))
		fmt.Printf("  ETag: %s\n", objInfo.ETag)
	}

	// 示例5: 下载文件
	fmt.Println("\n=== 下载文件 ===")
	reader, err := client.GetObject(ctx, bucketName, "test-file")
	if err != nil {
		log.Printf("下载文件失败: %v", err)
	} else {
		defer reader.Close()
		content := make([]byte, 1024)
		n, _ := reader.Read(content)
		fmt.Printf("文件内容: %s\n", string(content[:n]))
	}

	// 示例6: 列出存储桶中的对象
	fmt.Println("\n=== 列出存储桶中的对象 ===")
	objects, err := client.ListObjects(ctx, bucketName, "")
	if err != nil {
		log.Printf("列出对象失败: %v", err)
	} else {
		fmt.Printf("存储桶 %s 中的对象:\n", bucketName)
		for _, obj := range objects {
			fmt.Printf("  - %s (%s, %d bytes)\n", obj.Key, obj.ContentType, obj.Size)
		}
	}

	// 示例7: 带前缀的对象列表
	fmt.Println("\n=== 带前缀的对象列表 ===")
	objects, err = client.ListObjects(ctx, bucketName, "test")
	if err != nil {
		log.Printf("列出对象失败: %v", err)
	} else {
		fmt.Printf("前缀为 'test' 的对象:\n")
		for _, obj := range objects {
			fmt.Printf("  - %s\n", obj.Key)
		}
	}

	// 示例8: 范围下载
	fmt.Println("\n=== 范围下载 ===")
	rangeReader, err := client.GetObjectRange(ctx, bucketName, "test-file", 0, 5)
	if err != nil {
		log.Printf("范围下载失败: %v", err)
	} else {
		defer rangeReader.Close()
		content := make([]byte, 10)
		n, _ := rangeReader.Read(content)
		fmt.Printf("文件前6个字节: %s\n", string(content[:n]))
	}

	// 示例9: 获取直接访问URL
	fmt.Println("\n=== 获取直接访问URL ===")
	directURL := client.GetObjectURL(bucketName, "test-file")
	fmt.Printf("直接访问URL: %s\n", directURL)

	// 示例10: 删除对象
	fmt.Println("\n=== 删除对象 ===")
	if err := client.DeleteObject(ctx, bucketName, "test-file"); err != nil {
		log.Printf("删除对象失败: %v", err)
	} else {
		fmt.Println("对象删除成功")
	}

	// 示例11: 删除存储桶
	fmt.Println("\n=== 删除存储桶 ===")
	if err := client.DeleteBucket(ctx, bucketName); err != nil {
		log.Printf("删除存储桶失败: %v", err)
	} else {
		fmt.Printf("存储桶 %s 删除成功\n", bucketName)
	}
}

// 高级使用示例
func advancedExamples() {
	client := client.NewClient(client.ClientOptions{
		BaseURL: "http://localhost:8080",
		APIKey:  "your-api-key",
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// 批量上传文件
	fmt.Println("=== 批量上传文件 ===")
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	bucketName := "batch-upload-bucket"

	// 创建存储桶
	client.MakeBucket(ctx, bucketName)

	for i, filename := range files {
		content := strings.NewReader(fmt.Sprintf("Content of file %d", i+1))
		result, err := client.PutObject(ctx, bucketName, filename, content, filename)
		if err != nil {
			log.Printf("上传 %s 失败: %v", filename, err)
		} else {
			fmt.Printf("上传 %s 成功: %s\n", filename, result.Key)
		}
	}

	// 批量下载文件
	fmt.Println("\n=== 批量下载文件 ===")
	objects, err := client.ListObjects(ctx, bucketName, "")
	if err != nil {
		log.Printf("列出对象失败: %v", err)
		return
	}

	for _, obj := range objects {
		reader, err := client.GetObject(ctx, bucketName, obj.Key)
		if err != nil {
			log.Printf("下载 %s 失败: %v", obj.Key, err)
			continue
		}

		// 这里可以将内容写入本地文件
		content := make([]byte, obj.Size)
		n, _ := reader.Read(content)
		fmt.Printf("下载 %s: %s\n", obj.Key, string(content[:n]))
		reader.Close()
	}
}

// 错误处理示例
func errorHandlingExamples() {
	client := client.NewClient(client.ClientOptions{
		BaseURL: "http://localhost:8080",
		APIKey:  "invalid-key",
		Timeout: 5 * time.Second,
	})

	ctx := context.Background()

	// 尝试访问不存在的对象
	fmt.Println("=== 错误处理示例 ===")
	_, err := client.GetObject(ctx, "non-existent-bucket", "non-existent-key")
	if err != nil {
		fmt.Printf("期望的错误: %v\n", err)
	}

	// 超时处理
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err = client.ListObjects(ctx, "test-bucket", "")
	if err != nil {
		fmt.Printf("超时错误: %v\n", err)
	}
}
