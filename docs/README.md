# GF-File-Tool
A lightweight, cross-platform file compression/encryption tool written in Go, supporting zip/tar.gz compression, split compression, AES/DES encryption, and batch processing.

Also a training program from Chengdu University of Information Technology (CUIT). The aim is to cultivate Golang engineers.

Go语言CLI工具开发教学案例, 通过这个案例你可以学习到 Cobra 框架的基本用法, Viper 配置管理库的基本用法, 大量的 IO 读写训练, 文件/协议头部的解析, 密码学的一些基础知识. 你可以尝试修复该工具中一些显而易见的错误或是优化和新增更多的相关命令, 适合在学习完 Go 基础后配套使用, 祝你早日成为一名合格的 Golang 软件工程师.

## Features
✅ **Multi-format Compression**: Support zip/tar.gz compression/decompression  
✅ **Split Compression**: Split large files into small parts (zip only)  
✅ **Multi-algorithm Encryption**: AES-256/DES encryption for files  
✅ **Batch Processing**: Compress/encrypt multiple files/directories at once  
✅ **Cross-platform**: Support Windows/Linux (binary files in `bin/` directory)  
✅ **Progress Bar**: Real-time progress display for large file processing

## Quick Start
### 1. Download Binary
- Windows: `bin/gf-file-tool.exe`
- Linux: `bin/gf-file-tool`

### 2. Core Commands
#### Compress a single file (zip)
```powershell
./gf-file-tool.exe compress ./test/data/big-file.txt -o ./test/output/big-file.zip --verbose
```

#### Encrypt a compressed file (DES)
```powershell
./gf-file-tool.exe encrypt ./test/output/big-file.zip -k 12345678 -a des -o ./test/output/big-file.zip.des.enc --verbose
```

#### Decrypt and decompress
```powershell
./gf-file-tool.exe decrypt ./test/output/big-file.zip.des.enc -k 12345678 -a des -s <GENERATED_SALT> -o ./test/output/big-file-dec.zip --verbose
./gf-file-tool.exe decompress ./test/output/big-file-dec.zip -o ./test/output/decompress --verbose
```

## Project Structure
```plaintext
gf-file-tool/
├── cmd/          # Command-line interface (CLI) commands
├── core/         # Core logic (compression/crypto)
├── utils/        # Utility functions (file/key/salt handling)
├── progress/     # Progress bar implementation
├── test/         # Test data and output
└── docs/         # Documentation
```