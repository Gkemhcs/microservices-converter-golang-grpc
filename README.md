# Microservices Converter Demo

## Overview

This project is a demonstration of a microservices architecture implemented in Go (Golang). The system consists of several services that perform various media conversion tasks, such as converting video to audio, text to speech, and images to PDFs. Each service is designed to be independent, scalable, and easy to maintain.

## Services

### Frontend
- **Description**: The frontend service provides a user interface for interacting with the various conversion services.
- **Port**: 8080
- **Technologies**: Go, Gin, HTML, CSS, JavaScript

### Text-to-Speech
- **Description**: This service converts text input into speech audio files.
- **Port**: 8081
- **Technologies**: Go, Google Text-to-Speech API

### Video-to-Audio
- **Description**: This service extracts audio from video files and converts it into MP3 format.
- **Port**: 8082
- **Technologies**: Go, FFmpeg

### Image-to-PDF
- **Description**: This service converts image files into PDF documents.
- **Port**: 8083
- **Technologies**: Go, ImageMagick

### File Uploader
- **Description**: This service handles the uploading of files to Google Cloud Storage and generates signed URLs for accessing the files.
- **Port**: 8084
- **Technologies**: Go, Google Cloud Storage

## Technology Stack

- **Programming Language**: Go (Golang)
- **API Gateway**: Gin
- **Message Broker**: N/A
- **Database**: PostgreSQL
- **Cache**: Redis
- **Tracing**: OpenTelemetry
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Cloud Storage**: Google Cloud Storage

## Setup and Installation

### Prerequisites

- Docker
- Docker Compose
- Google Cloud Service Account Key (JSON file)

### Steps

1. **Clone the repository**:
    ```sh
    git clone https://github.com/yourusername/converter-gcp.git
    cd converter-gcp
    ```

2. **Set up environment variables**:
    Ensure you have a `key.json` file for Google Cloud credentials and place it in the project root directory.

3. **Build and run the services**:
    ```sh
    docker-compose up --build
    ```

4. **Access the services**:
    - Frontend: `http://localhost:8080`
    - Text-to-Speech: `http://localhost:8081`
    - Video-to-Audio: `http://localhost:8082`
    - Image-to-PDF: `http://localhost:8083`
    - File Uploader: `http://localhost:8084`

## Usage

1. **Frontend**:
    - Open the frontend URL in your browser.
    - Use the provided interface to upload files and perform conversions.

2. **API Endpoints**:
    - Each service exposes gRPC endpoints for performing conversions. Refer to the respective service's proto files for detailed API documentation.

## gRPC Explanation

gRPC is a high-performance, open-source universal RPC framework initially developed by Google. It uses HTTP/2 for transport, Protocol Buffers as the interface description language, and provides features such as authentication, load balancing, and more.

### How gRPC is Used in This Project

Each microservice in this project exposes a gRPC API for performing its specific conversion task. The gRPC services are defined using Protocol Buffers (proto files), which provide a language-agnostic way to define the service interfaces and message types.

### Example: Video-to-Audio Service

The `Video-to-Audio` service exposes a gRPC endpoint for converting video files to audio. The service definition in the proto file might look like this:

```proto
syntax = "proto3";

package converter;

service VideoToAudioConverterService {
  rpc Convert(stream VideoChunk) returns (ConvertVideoToAudioResponse);
}

message VideoChunk {
  bytes chunk = 1;
}

message ConvertVideoToAudioResponse {
  string url = 1;
}
```

### Making a gRPC Request

To make a gRPC request to the `Video-to-Audio` service, you would use a gRPC client in your preferred programming language. Here is an example in Go:

```go
package main

import (
  "context"
  "log"
  "os"
  "time"

  pb "converter/video-to-audio/genproto"
  "google.golang.org/grpc"
)

func main() {
  conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
  if err != nil {
    log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()

  client := pb.NewVideoToAudioConverterServiceClient(conn)

  stream, err := client.Convert(context.Background())
  if err != nil {
    log.Fatalf("could not convert: %v", err)
  }

  // Read video file and send chunks
  file, err := os.Open("video.mp4")
  if err != nil {
    log.Fatalf("could not open file: %v", err)
  }
  defer file.Close()

  buffer := make([]byte, 64*1024)
  for {
    n, err := file.Read(buffer)
    if err == io.EOF {
      break
    }
    if err != nil {
      log.Fatalf("could not read file: %v", err)
    }

    if err := stream.Send(&pb.VideoChunk{Chunk: buffer[:n]}); err != nil {
      log.Fatalf("could not send chunk: %v", err)
    }
  }

  response, err := stream.CloseAndRecv()
  if err != nil {
    log.Fatalf("could not receive response: %v", err)
  }

  log.Printf("Conversion completed, file URL: %s", response.GetUrl())
}
```

## Contributing

We welcome contributions to improve the project. Please fork the repository and submit pull requests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Gin](https://github.com/gin-gonic/gin)
- [FFmpeg](https://ffmpeg.org/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Google Cloud Storage](https://cloud.google.com/storage)
