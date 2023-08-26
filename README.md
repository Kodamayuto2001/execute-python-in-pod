# execute-python-in-pod

# インストール
```sh
go get k8s.io/client-go@latest
go get k8s.io/apimachinery@latest

go mod tidy
```

# 実行（数秒まつ）
```sh
$ go run main.go                          6.4s  Sun Aug 27 03:02:31 2023
Pod logs:
 hello, world
```

