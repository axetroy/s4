#!/bin/bash

# Reference:
# https://github.com/golang/go/blob/master/src/go/build/syslist.go
os_archs=(
    darwin/amd64
    linux/amd64
    windows/amd64
)

releases=()

for os_arch in "${os_archs[@]}"
do
    goos=${os_arch%/*}
    goarch=${os_arch#*/}

    filename=s4

    if [[ ${goos} == "windows" ]];
    then
        filename+=.exe
    fi

    echo building ${os_arch}

    CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build -ldflags "-s -w" -o ./bin/${filename} main.go >/dev/null 2>&1

    # if build success
    if [[ $? == 0 ]];then
        releases+=(${os_arch})
        cd ./bin

        tar -czf s4_${goos}_${goarch}.tar.gz ${filename}

        rm -rf ./${filename}

        cd ../
    fi
done

echo "release:"

for os_arch in "${releases[@]}"
do
    printf "\t%s\n" "${os_arch}"
done
echo