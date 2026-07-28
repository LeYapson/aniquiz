package handlers

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
)

// SendPasswordResetEmail envoie un email de réinitialisation de mot de passe via SMTP.
// Variables d'env requises : SMTP_HOST, SMTP_FROM, SMTP_USERNAME, SMTP_PASSWORD.
// SMTP_PORT est optionnel (défaut : 587). APP_URL est optionnel (défaut : https://aniquiz.fr).
func SendPasswordResetEmail(toEmail, username, resetURL string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")

	if host == "" || from == "" {
		return fmt.Errorf("SMTP non configuré (SMTP_HOST et SMTP_FROM requis)")
	}
	if port == "" {
		port = "587"
	}

	subject := "Réinitialisation de mot de passe — AniQuiz"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8">
  <style>
    body{font-family:Arial,sans-serif;background:#0f0f23;color:#f1f5f9;margin:0;padding:0}
    .wrap{max-width:480px;margin:40px auto;background:#1a1a3e;border-radius:12px;padding:32px}
    h1{color:#818cf8;font-size:1.4rem;margin-bottom:16px}
    p{color:#cbd5e1;line-height:1.6;margin:8px 0}
    .btn{display:inline-block;background:#6366f1;color:#ffffff !important;text-decoration:none;padding:12px 28px;border-radius:8px;font-weight:bold;margin:20px 0;font-size:1rem}
    .url{word-break:break-all;color:#818cf8;font-size:.8rem}
    .footer{color:#64748b;font-size:.78rem;margin-top:24px;border-top:1px solid #2d2d5e;padding-top:16px}
  </style>
</head>
<body>
  <div class="wrap">
    <h1>🔑 Réinitialisation de mot de passe</h1>
    <p>Bonjour <strong>%s</strong>,</p>
    <p>Tu as demandé à réinitialiser ton mot de passe sur <strong>AniQuiz</strong>.</p>
    <p>Clique sur le bouton ci-dessous pour créer un nouveau mot de passe.<br>Ce lien expire dans <strong>1 heure</strong>.</p>
    <a href="%s" class="btn">Réinitialiser mon mot de passe</a>
    <p>Si le bouton ne fonctionne pas, copie ce lien dans ton navigateur :</p>
    <p class="url">%s</p>
    <p>Si tu n'as pas fait cette demande, ignore cet email. Ton mot de passe restera inchangé.</p>
    <p class="footer">— L'équipe AniQuiz · <a href="https://aniquiz.fr" style="color:#818cf8">aniquiz.fr</a></p>
  </div>
</body>
</html>`, username, resetURL, resetURL)

	msg := "From: AniQuiz <" + from + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + htmlBody

	addr := net.JoinHostPort(host, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connexion SMTP échouée : %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("client SMTP : %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("STARTTLS : %w", err)
		}
	}

	if user != "" && pass != "" {
		if err = client.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
			return fmt.Errorf("auth SMTP : %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(toEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = fmt.Fprint(w, msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
