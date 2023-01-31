package main

import (
	"bytes"
	"crypto/tls"
	"gopkg.in/gomail.v2"
	"io"
	"log"
	"mime/multipart"
	"strings"
)

type MailerConfig struct {
	MailerSMTPHost    string
	MailerSMTPPort    int
	MailerAddressFrom string
	MailerPassword    string
}

func SendMail(sendTo []string, subject string, body string, attachment map[string]multipart.File, config MailerConfig) (err error) {

	msg := gomail.NewMessage()
	msg.SetHeader("To", strings.Join(sendTo, ","))
	msg.SetHeader("From", config.MailerAddressFrom)
	msg.SetHeader("Return-path", config.MailerAddressFrom)
	msg.SetHeader("Subject", subject)

	msg.SetBody("text/html", body)
	if len(attachment) > 0 {
		for filename, file := range attachment {
			buf := bytes.NewBuffer(nil)
			_, err := io.Copy(buf, file)
			if err != nil {
				continue
			}
			msg.Attach(
				filename,
				gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := w.Write(buf.Bytes())
					if err != nil {
						return err
					}
					return nil
				},
				),
				gomail.SetHeader(map[string][]string{
					"Content-Type":              {"text/plain; charset=\"utf-8\""},
					"Content-Transfer-Encoding": {"base64"},
					"Content-Disposition":       {"attachment;filename=\"" + filename + "\""},
				},
				),
			)
		}
	}

	dialer := gomail.NewDialer(
		config.MailerSMTPHost,
		config.MailerSMTPPort,
		config.MailerAddressFrom,
		config.MailerPassword,
	)

	//из-за ошибки x509: certificate is valid for *.from.sh, from.sh, not smtp.securige.com
	//отключил проверку TLS, ибо судя по всему на sprinthost'е коряво настроены сертификаты,
	//поэтому же используется gomail вместо стандартного net/smtp
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err = dialer.DialAndSend(msg)
	if err != nil {
		log.Printf("SendMail error: %v\n", err)
		return err
	}
	return nil
}
