// Command unpack 把 v2ray 格式的 geosite.dat 解包成 domain-list-community 能识别的纯文本列表。
// 每个分组写一个文件，文件名即分组名（dat 里的 category），
// 每行一条规则：domain: / full: / keyword: / regexp:。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

func main() {
	inputData := flag.String("inputdata", "geosite.dat", "geosite.dat 文件路径")
	outputDir := flag.String("outputdir", "./data", "纯文本列表输出目录")
	flag.Parse()

	data, err := os.ReadFile(*inputData)
	if err != nil {
		fmt.Printf("[Fatal] 读取 %s 失败: %v\n", *inputData, err)
		os.Exit(1)
	}
	siteList := new(router.GeoSiteList)
	if err := proto.Unmarshal(data, siteList); err != nil {
		fmt.Printf("[Fatal] 解析 geosite.dat 失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("[Fatal] 创建输出目录失败: %v\n", err)
		os.Exit(1)
	}
	for _, site := range siteList.Entry {
		f, err := os.Create(filepath.Join(*outputDir, strings.ToLower(site.CountryCode)))
		if err != nil {
			fmt.Printf("[Fatal] 创建 %s 失败: %v\n", site.CountryCode, err)
			os.Exit(1)
		}
		w := bufio.NewWriter(f)
		for _, d := range site.Domain {
			var typ string
			switch d.Type {
			case router.Domain_RootDomain:
				typ = "domain"
			case router.Domain_Full:
				typ = "full"
			case router.Domain_Plain:
				typ = "keyword"
			case router.Domain_Regex:
				typ = "regexp"
			default:
				fmt.Printf("[Fatal] %s 里出现未知规则类型 %v\n", site.CountryCode, d.Type)
				os.Exit(1)
			}
			w.WriteString(typ)
			w.WriteByte(':')
			w.WriteString(d.Value)
			for _, attr := range d.Attribute {
				w.WriteByte(' ')
				w.WriteByte('@')
				w.WriteString(attr.Key)
			}
			w.WriteByte('\n')
		}
		if err := w.Flush(); err != nil {
			fmt.Printf("[Fatal] 写入 %s 失败: %v\n", site.CountryCode, err)
			os.Exit(1)
		}
		f.Close()
		fmt.Printf("分组 %s: %d 条规则\n", site.CountryCode, len(site.Domain))
	}
}
