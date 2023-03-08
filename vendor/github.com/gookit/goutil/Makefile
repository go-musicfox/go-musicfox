# link https://github.com/humbug/box/blob/master/Makefile
#SHELL = /bin/sh
.DEFAULT_GOAL := help
# 每行命令之前必须有一个tab键。如果想用其他键，可以用内置变量.RECIPEPREFIX 声明
# mac 下这条声明 没起作用 !!
#.RECIPEPREFIX = >
.PHONY: all usage help clean

# 需要注意的是，每行命令在一个单独的shell中执行。这些Shell之间没有继承关系。
# - 解决办法是将两行命令写在一行，中间用分号分隔。
# - 或者在换行符前加反斜杠转义 \

# 接收命令行传入参数 make COMMAND tag=v2.0.4
# TAG=$(tag)

##There some make command for the project
##

help:
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//' | sed -e 's/: / /'

##Available Commands:

  readme:     ## Generate or update README file by ./internal/gendoc
readme:
	go run ./internal/gendoc -o README.md
	go run ./internal/gendoc -o README.zh-CN.md -l zh-CN

  readme-c:     ## Generate or update README file and commit change to git
readme-c: readme
	git add README.* internal
	git commit -m "doc: update and re-generate README docs"

  csfix:      ## Fix code style for all files by go fmt
csfix:
	go fmt ./...

  csdiff:     ## Display code style error files by gofmt
csdiff:
	gofmt -l ./
