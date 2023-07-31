package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/peacecwz/go-mac-imessage/config"
	"github.com/peacecwz/go-mac-imessage/global"
	"github.com/peacecwz/go-mac-imessage/sms"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

func main() {
	log.Println("读取配置文件")
	global.Config = config.Config{}
	config, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("读取配置文件失败", err.Error())
		return
	}

	err = yaml.Unmarshal(config, &global.Config)
	if err != nil {
		log.Println("解析配置文件失败", err.Error())
		return
	}
	log.Println("配置文件加载成功")

	log.Println("初始化OpenAI")
	global.AIClient = openai.NewClient(global.Config.OpenAI.GPTToken)
	global.Context = context.Background()
	log.Println("OpenAI初始化成功")

	log.Println("开始等待iMessage短信...")
	var interval int64 = 1000
	err = sms.TrackSMS(interval, func(smss []sms.SMS) {
		for _, s := range smss {

			if s.IsRead {
				continue
			}

			// 接收iMessage信息并解析
			log.Printf("接收到短信: %s 发送自 %s\n", s.Content, s.From)

			if len(s.From) > 11+3 {
				log.Println("发送者号码过长，跳过")
				continue
			}

			err := s.Read()
			if err != nil {
				log.Println("读取信息失败", err.Error())
			}

			log.Println("开始请求ChatGPT")

			// 请求ChatGPT
			resp, err := global.AIClient.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT3Dot5Turbo,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: s.Content,
						},
					},
				},
			)

			if err != nil {
				msg := fmt.Sprintf("请求GPT失败: %s", err.Error())
				msg = strings.ReplaceAll(msg, "\"", "")
				msg = strings.ReplaceAll(msg, "'", "")

				log.Println(msg)

				err = sms.Send(msg, s.From)
				if err != nil {
					log.Println("发送信息失败", err.Error())
					continue
				}

				continue
			}

			// 发送GPT回答
			err = sms.Send(resp.Choices[0].Message.Content, s.From)
			if err != nil {
				log.Println("发送信息失败", err.Error())
				continue
			}
		}
	})

	if err != nil {
		log.Println("接收信息失败", err.Error())
		return
	}
}
