package utils

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"
)

// SendVerificationEmail 发送验证码邮件
func SendVerificationEmail(email, code string) error {
	// 从环境变量获取邮件配置
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpServer := os.Getenv("EMAIL_SMTP_SERVER")
	smtpPort := os.Getenv("EMAIL_SMTP_PORT")
	if from == "" || password == "" || smtpServer == "" || smtpPort == "" {
		log.Println("SMTP configuration is missing")
		return fmt.Errorf("SMTP configuration is missing")
	}

	if from == "" || password == "" || smtpServer == "" || smtpPort == "" {
		log.Println("SMTP configuration is missing")
		return fmt.Errorf("SMTP configuration is missing")
	}

	to := []string{email}
	subject := "Email Verification"
	body := fmt.Sprintf("Your verification code is: %s", code)

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		ServerName: smtpServer,
	}

	// 连接到 SMTP 服务器
	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
	if err != nil {
		log.Printf("Failed to connect to SMTP server: %v", err)
		return err
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		log.Printf("Failed to create SMTP client: %v", err)
		return err
	}
	defer client.Quit()

	// 进行认证
	auth := smtp.PlainAuth("", from, password, smtpServer)
	if err = client.Auth(auth); err != nil {
		log.Printf("Failed to authenticate: %v", err)
		return err
	}

	// 设置发件人和收件人
	if err = client.Mail(from); err != nil {
		log.Printf("Failed to set sender: %v", err)
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			log.Printf("Failed to set recipient: %v", err)
			return err
		}
	}

	// 发送邮件数据
	writer, err := client.Data()
	if err != nil {
		log.Printf("Failed to start data transfer: %v", err)
		return err
	}
	_, err = writer.Write([]byte("To: " + email + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body))
	if err != nil {
		log.Printf("Failed to write message: %v", err)
		return err
	}
	err = writer.Close()
	if err != nil {
		log.Printf("Failed to close data transfer: %v", err)
		return err
	}

	log.Printf("Email sent to %s", email)
	return nil
}

// GenerateVerificationCode 生成验证码
func GenerateVerificationCode() string {
	// 定义验证码的最大值，即 999999
	max := big.NewInt(1000000)
	// 使用 crypto/rand 生成一个介于 0 到 999999 之间的随机整数
	num, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Printf("Failed to generate random number: %v", err)
		return ""
	}
	// 将随机整数格式化为 6 位字符串，不足 6 位时在前面补 0
	return fmt.Sprintf("%06d", num.Int64())
}
