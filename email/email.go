package email

import (
	"crypto/tls"
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"net/smtp"
	"strconv"
)

type EmailClient struct {
	Cfg  *config.Config
	auth smtp.Auth
}

func NewEmailClient(cfg *config.Config) *EmailClient {
	auth := smtp.PlainAuth("", cfg.Email.Login, cfg.Email.Password, cfg.Email.Host)
	client := &EmailClient{
		Cfg:  cfg,
		auth: auth,
	}
	return client
}

func (ec *EmailClient) SendEmail(recipient string, text string) error {
	to := []string{recipient}
	msg := []byte("To: " + recipient + "\r\n" + "Subject: Код авторизации amoCRM\r\n" + "\r\n" + text)
	err := smtp.SendMail(ec.Cfg.Email.Host+":"+strconv.Itoa(ec.Cfg.Email.Port),
		ec.auth,
		ec.Cfg.Email.Host,
		to,
		msg)
	if err != nil {
		return err
	}

	return nil
}
func (ec *EmailClient) SendEmailWithTLS(to, body string) error {
	host := ec.Cfg.Email.Host
	port := strconv.Itoa(ec.Cfg.Email.Port)
	address := host + ":" + port

	// Явно открываем TLS-соединение
	conn, err := tls.Dial("tcp", address, nil)
	if err != nil {
		return fmt.Errorf("failed to dial TLS: %w", err)
	}
	defer conn.Close()

	// Создаем SMTP-Client поверх TLS-соединения
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Производим Авторизацию на сервере
	if err := client.Auth(ec.auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	// Указываем отправителя(сообщаем серверу от кого будет письмо)
	if err := client.Mail(ec.Cfg.Email.Login); err != nil {

		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Указываем получателя (сообщаем серверу кому будет отправлено письмо)
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Начало передачи содержимого письма (Получаем io.WriteCloser, в который можно писать тело сообщения.)
	dataWriter, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	msg := []byte(
		"Subject: Код авторизации amoCRM\r\n" + "\r\n" + fmt.Sprintf("%s", body))

	// Запись письма на сервер (пишем тело письма в поток)
	if _, err := dataWriter.Write(msg); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	// Завершение отправки(Закрываем поток, сервер получает сигнал о том, что мы закончили отправку письма)
	if err := dataWriter.Close(); err != nil {
		return fmt.Errorf("failed to close data stream: %w", err)
	}
	// Письмо полностью отправлено

	return nil
}
