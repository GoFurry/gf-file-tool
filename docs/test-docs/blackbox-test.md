# 测试用例

## 1 准备测试目录

```
│      
├─bin
│      gf-file-tool.exe
│      
├─test
│  ├─data
│  │      big-file.txt
│  │      super-big-file.vpk
│  │      test1.txt
│  │      test2.txt
│  │      
│  └─output
```

## 2 黑盒测试

### 2.1.1 zip/tar.gz 简易压缩 

```cmd
.\bin\gf-file-tool.exe compress .\test\data\big-file.txt -o .\test\output\big-file.zip --verbose

.\bin\gf-file-tool.exe compress .\test\data\big-file.txt -o .\test\output\big-file.tar.gz --format targz --verbose
```

预期结果：output文件夹内生成big-file.zip、big-file.tar.gz, 两个压缩文件均无损坏.

### 2.1.2 zip/tar.gz 解压缩

```powershell
.\bin\gf-file-tool.exe decompress .\test\output\big-file.zip -o .\test\output\decompress\big-file --verbose

.\bin\gf-file-tool.exe decompress .\test\output\big-file.tar.gz -o .\test\output\decompress\big-file-tar -f targz --verbose  
```

预期结果：output文件夹内生成decompress\big-file和decompress\big-file-tar,提取出的被压缩文件无损坏.

### 2.1.3 zip/tar.gz 批量压缩

```powershell
.\bin\gf-file-tool.exe compress .\test\data\test1.txt .\test\data\test2.txt -o .\test\output\batch-files.zip --verbose

.\bin\gf-file-tool.exe compress .\test\data -o .\test\output\data-dir.tar.gz --format targz --verbose
```

预期结果：output文件夹内生成batch-files.zip和data-dir.tar.gz, 两个压缩文件均无损坏.

### 2.1.4 zip/tar.gz 批量解压缩

```powershell
.\bin\gf-file-tool.exe decompress .\test\output\batch-files.zip -o .\test\output\decompress\batch --verbose

.\bin\gf-file-tool.exe decompress .\test\output\data-dir.tar.gz -o .\test\output\decompress\data-dir -f targz --verbose
```

预期结果：output文件夹内生成decompress\batch和decompress\data-dir,提取出的被压缩文件无损坏.

### 2.2.1 zip 分卷压缩

```powershell
.\bin\gf-file-tool.exe compress .\test\data\super-big-file.vpk -o .\test\output\super-split.zip --split 200000000 --verbose
```

预期结果：output文件夹内生成分卷压缩文件, zip文件无损坏.

### 2.2.2 zip 分卷解压缩

```cmd
.\bin\gf-file-tool.exe decompress .\test\output\super-split.zip.001 -o .\test\output\decompress\super-split --verbose
```

预期结果：output文件夹内生成decompress\super-split,提取出的被压缩文件无损坏.

### 2.2.3 合并分卷压缩包

```cmd
.\bin\gf-file-tool.exe merge .\test\output\super-split.zip -o .\test\output\super-merged.zip --verbose
```

预期结果：output文件夹内生成super-merged.zip, zip文件无损坏.

### 2.2.4 zip 批量分卷压缩

```cmd
.\bin\gf-file-tool.exe compress .\test\data -o .\test\output\data-split.zip --split 200000000 --verbose
```

预期结果：output文件夹内生成分卷压缩文件, zip文件无损坏.

### 2.2.5 zip 批量分卷解压缩

```
.\bin\gf-file-tool.exe decompress .\test\output\data-split.zip.001 -o .\test\output\decompress\data-split --verbose
```

预期结果：output文件夹内生成decompress\data-split,提取出的被压缩文件无损坏.

### 2.3.1 zip 加密压缩

```powershell
.\bin\gf-file-tool.exe compress .\test\data\big-file.txt -o .\test\output\big-file-enc.zip -e -k 123456 --verbose
```

预期结果：output文件夹内生成big-file-enc.zip, zip文件内容为密文.

### 2.3.2 zip 解密解压缩

```cmd
.\bin\gf-file-tool.exe decompress .\test\output\big-file-enc.zip -o .\test\output\decompress\big-file-enc -e -k 123456 -l 32 --verbose
```

预期结果：output文件夹内生成decompress\big-file-enc, 解压后内容为明文.

### 2.3.3 zip 批量分卷加密压缩

```cmd
.\bin\gf-file-tool.exe compress .\test\data\super-big-file.vpk -o .\test\output\super-split-enc.zip --split 200000000 -e -k 12345678 --verbose
```

预期结果：output文件夹内生成密文分卷压缩文件, zip文件无损坏.

### 2.3.4 zip 批量分卷解密解压缩

```cmd
.\bin\gf-file-tool.exe decompress .\test\output\super-split-enc.zip.001 -o .\test\output\decompress\super-split-enc -e -k 12345678 -l 32 --verbose
```

预期结果：output文件夹内生成decompress\super-split-enc, 解压后内容为明文.

### 2.4.1 zip/tar.gz 简易压缩后文件进行 AES/DES 加密

```powershell
.\bin\gf-file-tool.exe encrypt .\test\output\big-file.zip -k 123456 -a aes -o .\test\output\big-file.zip.aes.enc --verbose

..\bin\gf-file-tool.exe encrypt .\test\output\big-file.tar.gz -k 12345678 -a des -o .\test\output\big-file.tar.gz.des.enc --verbose
```

预期结果：output文件夹内生成big-file.zip.aes.enc、big-file.tar.gz.des.enc两个加密文件.

### 2.4.2 zip/tar.gz 简易压缩后文件进行 AES/DES 解密

```cmd
# 2.4.2 Decrypt AES-encrypted zip file (replace <SALT> with generated salt)
.\bin\gf-file-tool.exe decrypt .\test\output\big-file.zip.aes.enc -k 123456 -a aes -s <SALT> -o .\test\output\big-file-dec.zip --verbose
# 2.4.2 Decrypt DES-encrypted tar.gz file (replace <SALT> with generated salt)
.\bin\gf-file-tool.exe decrypt .\test\output\big-file.tar.gz.des.enc -k 12345678 -a des -s <SALT> -o .\test\output\big-file-dec.tar.gz --verbose
```

预期结果：output文件夹内生成big-file-dec.zip、big-file-dec.tar.gz, zip文件无损坏，内容为明文.

### 2.5 合法性检测

```powershell
#对比md5值是否改变
Get-FileHash .\test\data\big-file.txt -Algorithm MD5
Get-FileHash .\test\output\decompress\big-file\big-file.txt -Algorithm MD5

Get-FileHash .\test\data\super-big-file.vpk -Algorithm MD5
Get-FileHash .\test\output\decompress\super-split\super-big-file.vpk -Algorithm MD5

Get-FileHash .\test\output\big-file.zip -Algorithm MD5
Get-FileHash .\test\output\big-file-dec.zip -Algorithm MD5
```