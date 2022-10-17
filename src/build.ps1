Remove-Item bin -Recurse -Force
go build -o bin\go-crawer.exe server.go
go build -o bin\admin.exe admin\admin.go
Copy-Item config\config.yaml bin\config.yaml
Copy-Item admin\admin_config.yaml bin\admin_config.yaml
Copy-Item parser\rules\ bin\rules\ -Recurse