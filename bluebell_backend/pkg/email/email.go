package email

import (
	"bluebell_backend/models"
	"bluebell_backend/settings"
	"crypto/tls"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2" // 发送电子邮件包
)

// SendEmail 发送邮件
func SendEmail(p *models.RegisterEmailData) error {
	zap.L().Debug("Sending email", zap.String("to", p.Email),
		zap.String("username", p.UserName))

	if p == nil || p.Email == "" || p.UserName == "" {
		return fmt.Errorf("invalid email data")
	}

	if settings.Conf == nil || settings.Conf.EmailConfig == nil {
		return fmt.Errorf("email configuration is not initialized")
	}

	message := `
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <style>
        p {
            margin: 1em 0;
            text-indent: 2em;
            font-family: Arial, 'Microsoft YaHei', sans-serif;
        }
    </style>
</head>
<body>
    <h1>%s，欢迎加入Bluebell 编程博客论坛！</h1>
    
    <p>感谢您加入我们的社区，在这里，充满热情的人们聚集在一起，讨论和分享各种话题。</p>
    
    <p>我们很高兴您成为我们不断增长的社区的一部分。现在，您可以参与讨论，开新帖子，并与其他成员分享您的见解。</p>

    <p>下面是一些帮助您快速开始的小提示：
        <ul>
            <li><strong>探索主题：</strong>浏览不同的版块，找到您感兴趣的话题。</li>
            <li><strong>参与社区：</strong>通过分享您的想法和经验，为讨论做出贡献。</li>
            <li><strong>保持更新：</strong>开启通知，跟上最新的帖子和回复。</li>
        </ul>
    </p>

    <p>请务必熟悉我们的<a href="URL_TO_COMMUNITY_GUIDELINES">社区准则</a>，以确保每位参与者都能获得积极的体验。</p>

    <p>如果您有任何问题或需要帮助，随时联系我们的支持团队。</p>
    
    <p>祝好，<br> Bluebell编程博客论坛团队</p>
</body>
</html>
`

	host := "smtp.qq.com" // SMTP服务器
	port := 465           // 使用SSL协议端口
	username := settings.Conf.EmailConfig.UserName
	password := settings.Conf.EmailConfig.Password

	if username == "" || password == "" {
		return fmt.Errorf("email configuration is missing")
	}

	zap.L().Debug("Sending email", zap.String("to", p.Email),
		zap.String("username", p.UserName))

	m := gomail.NewMessage() // 创建Message实例
	// 设置发件人、收件人、邮件主题、邮件内容
	m.SetHeader("From", username)
	m.SetHeader("To", p.Email)
	m.SetHeader("Subject", "Welcome to Bluebell Forum!")
	m.SetBody("text/html", fmt.Sprintf(message, p.UserName))

	d := gomail.NewDialer(
		host,
		port,
		username,
		password,
	)
	// 跳过证书验证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
