# Restore objects from S3 Glacier Deep Archive

It iterates through objects at the specified S3 path, identifies objects in
Deep Archive, and initiates a restoration request for them if they are not
already restored or in the process of being restored.


## Build

```powershell
go build -ldflags="-s -w" -trimpath -o icebreaker.exe main.go
move-item -force ./icebreaker.exe ~/.local/bin/
```

## Run

```powershell
$env:AWS_PROFILE="xxx"
icebreaker -path s3://mybucket/myfolder
```
