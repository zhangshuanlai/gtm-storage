# gtm-storage
gtm-storage-sdk-go

一个用于GTM对象存储服务的Go SDK，提供简单易用的API来管理存储桶和对象。

## 特性

- 🚀 简单易用的API接口
- 📦 完整的存储桶管理功能
- 📄 支持对象的上传、下载、删除操作
- 🔍 对象元数据查询
- 📋 对象列表和前缀过滤
- 🎯 支持范围下载
- 🖼️ 自动生成预览和缩略图URL
- ⚡ 上下文支持和超时控制
- 🛡️ 完善的错误处理

## 安装

```bash
go get github.com/your-org/gtm-storage-sdk
```

## 快速开始

### 创建客户端

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/your-org/gtm-storage-sdk/gtmstorage"
)

func main() {
    // 创建客户端
    client := gtmstorage.NewClient(gtmstorage.ClientOptions{
        BaseURL: "http://localhost:8080", // GTM存储服务地址
        APIKey:  "your-api-key",          // API密钥
        Timeout: 30 * time.Second,        // 请求超时时间
    })
    
    ctx := context.Background()
    
    // 现在可以使用客户端了
    fmt.Println("GTM Storage 客户端创建成功")
}
```

### 基本操作

#### 1. 存储桶管理

```go
// 创建存储桶
err := client.MakeBucket(ctx, "my-bucket")
if err != nil {
    log.Fatal(err)
}

// 删除存储桶
err = client.DeleteBucket(ctx, "my-bucket")
if err != nil {
    log.Fatal(err)
}
```

#### 2. 对象上传

```go
// 从字符串上传
content := strings.NewReader("Hello, World!")
result, err := client.PutObject(ctx, "my-bucket", "hello.txt", content, "hello.txt")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("上传成功: %s\n", result.Key)
fmt.Printf("预览URL: %s\n", result.PreviewURL)

// 从文件上传
file, err := os.Open("local-file.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

result, err = client.PutObject(ctx, "my-bucket", "photos/image", file, "local-file.jpg")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("图片预览URL: %s\n", result.PreviewURL)
fmt.Printf("缩略图URL: %s\n", result.ThumbnailURL)
```

#### 3. 对象下载

```go
// 下载对象
reader, err := client.GetObject(ctx, "my-bucket", "hello.txt")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// 读取内容
content, err := io.ReadAll(reader)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("文件内容: %s\n", string(content))

// 范围下载（下载文件的一部分）
rangeReader, err := client.GetObjectRange(ctx, "my-bucket", "large-file.txt", 0, 100)
if err != nil {
    log.Fatal(err)
}
defer rangeReader.Close()
```

#### 4. 对象元数据

```go
// 获取对象元数据
info, err := client.HeadObject(ctx, "my-bucket", "hello.txt")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("文件大小: %d bytes\n", info.Size)
fmt.Printf("内容类型: %s\n", info.ContentType)
fmt.Printf("最后修改: %s\n", info.LastModified.Format(time.RFC3339))
fmt.Printf("ETag: %s\n", info.ETag)
```

#### 5. 列出对象

```go
// 列出存储桶中的所有对象
objects, err := client.ListObjects(ctx, "my-bucket", "")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("存储桶中有 %d 个对象:\n", len(objects))
for _, obj := range objects {
    fmt.Printf("- %s (%d bytes)\n", obj.Key, obj.Size)
}

// 使用前缀过滤
photoObjects, err := client.ListObjects(ctx, "my-bucket", "photos/")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("照片文件夹中有 %d 个文件\n", len(photoObjects))
```

#### 6. 删除对象

```go
// 删除单个对象
err := client.DeleteObject(ctx, "my-bucket", "hello.txt")
if err != nil {
    log.Fatal(err)
}

fmt.Println("文件删除成功")