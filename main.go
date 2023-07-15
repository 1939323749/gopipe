package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopipe/utils"
	"io"
	"os"
)

func main() {
	showOrigin := flag.Bool("o", false, "Include the original text in the output")
	flag.Parse()
	var targetLang = "ZH"
	if len(flag.Args()) > 0 {
		targetLang = flag.Args()[0]
	}
	reader := bufio.NewReader(os.Stdin)
	var s []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			_, err := fmt.Fprint(os.Stderr)
			if err != nil {
				return
			}
			os.Exit(1)
		}
		if len(line) > 0 {
			s = append(s, line)
		}
		if err == io.EOF {
			break
		}
	}
	for i := 0; i < len(s); i++ {
		translated, err := utils.Translation("EN", targetLang, s[i])
		if err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Translation error: %v\n", err)
			if err != nil {
				return
			}
			os.Exit(1)
		}
		fmt.Printf("%s", translated)
		if *showOrigin {
			if s[i] != "\n" {
				_, _ = fmt.Printf("--> %s", s[i])
			}
		}
	}
}
